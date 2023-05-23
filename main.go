package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/GeoffreyDick/gogarin/api"
	"github.com/GeoffreyDick/gogarin/lib"
	m "github.com/GeoffreyDick/gogarin/model"
	"github.com/charmbracelet/log"
	"github.com/joho/godotenv"
)

var (
	token string
)

func init() {
	l := log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
		Prefix:          "🏗️ INIT_BOT",
	})

	// load .env file
	err := godotenv.Load(".env")
	if err != nil {
		l.Fatal("Error loading .env file")
	}

	token = os.Getenv("TOKEN")

	if token == "" {
		l.Fatal("TOKEN environment variable not set")
	}
}

func main() {
	c := api.NewClient(token)

	// TerminalBot actions.
	tb := NewTerminalBot(c)

	// Get agent, verify, and welcome.
	agent, err := tb.GetMyAgent()
	if err != nil {
		tb.logger.Fatal("Failed to get agent", "error", err)
	}
	tb.logger.Infof("Agent verified. Welcome %s", agent.Symbol)

	// AgentBot actions.
	ab := NewAgentBot(c, agent)

	// Get contracts.
	ab.logger.Info("Getting contracts...")
	contracts, err := ab.GetMyContracts()
	if err != nil {
		tb.logger.Fatal("Failed to get contracts", "error", err)
	}
	ab.logger.Info("Contracts retrieved.", "count", len(*contracts))

	// Accept contracts if not already accepted.
	for _, contract := range *contracts {
		if !contract.Accepted {
			ab.logger.Info("Found new contract. Accepting...", "id", contract.ID)
			contract, err := c.AcceptContract(contract.ID)
			if err != nil {
				tb.logger.Fatal("Failed to accept contract", "error", err)
			}
			ab.logger.Info("Contract accepted.", "terms", contract.Contract.Terms)
		}
	}

	// Determine priorities.
	ab.logger.Info("Determining priorities...")
	priorities, err := ab.DeterminePriorities(contracts)
	if err != nil {
		tb.logger.Fatal("Failed to determine priorities", "error", err)
	}
	ab.logger.Info("Priorities determined.", "priorities", priorities)

	// sbCh contains a ShipBot for each ship in the fleet.
	// ShipBots sent to sbCh will be processed by the command loop.
	sbCh := make(chan ShipBot)
	done := make(chan bool)
	completed := make(chan bool)

	wg := sync.WaitGroup{}

	// Start ShipBot command loop.
	go func() {
		ab.logger.Info("Starting command loop...")
		for {
			select {
			case sb := <-sbCh:
				sb.logger.Info("Reporting in.", "role", sb.ship.Registration.Role)
				// RoleSwitch
				switch sb.ship.Registration.Role {
				case "COMMAND":
					// TODO: command ship logic
				case "EXCAVATOR":
					if sb.IsFullOfCargo() && sb.IsAtWaypointWithTrait("MARKETPLACE") && sb.ship.Nav.Status == "DOCKED" {
						ab.logger.Info(fmt.Sprintf("%s %s", sb.ship.Registration.Role, sb.ship.Symbol), "mission", "Sell cargo")
						go sb.SellCargo(sbCh)
					}

					if sb.IsFullOfCargo() && sb.IsAtWaypointWithTrait("MARKETPLACE") && sb.ship.Nav.Status != "DOCKED" {
						ab.logger.Info(fmt.Sprintf("%s %s", sb.ship.Registration.Role, sb.ship.Symbol), "mission", "Dock ship")
						go sb.DockShip(sbCh)
					}

					if sb.IsFullOfCargo() && !sb.IsAtWaypointWithTrait("MARKETPLACE") {
						ab.logger.Info(fmt.Sprintf("%s %s", sb.ship.Registration.Role, sb.ship.Symbol), "mission", "Navigate to nearest marketplace")
						go sb.NavigateToNearestWaypointWithTrait("MARKETPLACE", sbCh)
					}

					if !sb.IsFullOfCargo() && sb.IsAtWaypointOfType("ASTEROID_FIELD") {
						ab.logger.Info(fmt.Sprintf("%s %s", sb.ship.Registration.Role, sb.ship.Symbol), "mission", "Extract resources")
						go sb.ExtractResources(sbCh)
					}

					if !sb.IsFullOfCargo() && !sb.IsAtWaypointOfType("ASTEROID_FIELD") {
						ab.logger.Info(fmt.Sprintf("%s %s", sb.ship.Registration.Role, sb.ship.Symbol), "mission", "Navigate to nearest asteroid field")
						go sb.NavigateToNearestWaypointOfType("ASTEROID_FIELD", sbCh)
					}
				}
			case <-done:
				fmt.Println("exiting...")
				completed <- true
				return
			}
		}
	}()

	// Get fleet.
	ab.logger.Info("Waking fleet...")
	ships, err := c.GetMyShips()
	if err != nil {
		ab.logger.Fatal("Failed to get ships", "error", err)
	}

	// If only one ship, InitiateRequisitionProtocol.
	if len(*ships) > 0 {
		ab.logger.Info("Found only one ship. Sending command ship on requisition mission...")

		// InitiateRequisitionProtocol.
		ship := (*ships)[0]
		sb := NewShipBot(c, &ship, ab.agent)

		wg.Add(1)

		go sb.InitiateRequisitionProtocol(&wg)

		wg.Wait()
	}

	// Get fleet underway.
	for i, ship := range *ships {
		wg.Add(1)
		defer wg.Done()
		ship = (*ships)[i]

		go func(ship m.Ship) {
			// Create ShipBot.
			sb := NewShipBot(c, &ship, ab.agent)

			// Check if ship on cooldown
			sb.logger.Info("⚛ Checking reactor...")
			cooldown, err := sb.GetShipCooldown()
			if err != nil {
				sb.logger.Error("⚛ Error getting ship cooldown.", "error", err)
			}
			sb.cooldown = cooldown

			// Send sb to sbCh.
			sbCh <- *sb
		}(ship)
	}

	wg.Wait()

	// close(done)
	// <-completed
}

/*
🖥️ TERMINAL_BOT
*/

// TerminalBot represents a TerminalBot instance.
type TerminalBot struct {
	client *api.Client
	logger *log.Logger
}

// NewTerminalBot creates a new instance of TerminalBot.
func NewTerminalBot(c *api.Client) *TerminalBot {
	return &TerminalBot{
		client: c,
		logger: log.NewWithOptions(os.Stderr, log.Options{
			ReportTimestamp: true,
			Prefix:          "🖥️ TERMINAL_BOT",
		}),
	}
}

// GetMyAgent verifies an agent.
func (tb *TerminalBot) GetMyAgent() (*m.Agent, error) {
	tb.logger.Info("Credentials received. Retrieving agent...")
	agent, err := tb.client.GetMyAgent()
	if err != nil {
		return nil, err
	}

	return agent, nil
}

/*
👽 AGENT_BOT
*/

// AgentBot represents an AgentBot instance.
type AgentBot struct {
	client     *api.Client
	logger     *log.Logger
	agent      *m.Agent
	contracts  *[]m.Contract
	priorities *[]string
}

// NewAgentBot creates a new instance of AgentBot.
func NewAgentBot(client *api.Client, agent *m.Agent) *AgentBot {
	return &AgentBot{
		client: client,
		logger: log.NewWithOptions(os.Stderr, log.Options{
			ReportTimestamp: true,
			Prefix:          fmt.Sprintf("👽 %s", agent.Symbol),
		}),
		agent: agent,
	}
}

// GetMyContracts retrieves the Agent's contracts.
func (ab *AgentBot) GetMyContracts() (*[]m.Contract, error) {
	contracts, err := ab.client.GetMyContracts()
	if err != nil {
		return nil, err
	}

	return contracts, nil
}

// SetPriorities scrapes the agent's contracts for priority trade goods.
func (ab *AgentBot) DeterminePriorities(contracts *[]m.Contract) (*[]string, error) {
	var priorities []string

	for _, contract := range *contracts {
		for _, good := range contract.Terms.Deliver {
			if !lib.Contains(priorities, good.TradeSymbol) {
				priorities = append(priorities, good.TradeSymbol)
			}
		}
	}

	return &priorities, nil
}

/*
🚀 SHIP_BOT
*/

// ShipBot represents a ShipBot instance.
type ShipBot struct {
	client     *api.Client
	logger     *log.Logger
	agent      *m.Agent
	contracts  *[]m.Contract
	priorities []string
	ship       *m.Ship
	cooldown   *m.Cooldown
}

// NavigateToNearestWaypointOfType: Navigate to nearest waypoint of type.
func (sb *ShipBot) NavigateToNearestWaypointOfType(waypointType string, sbCh chan ShipBot) {
	sb.logger.Info("Navigating to nearest waypoint of type...", "waypointType", waypointType)

	// Get nearest waypoint of type.
	waypoints, err := sb.client.ListWaypoints(sb.ship.Nav.SystemSymbol)
	if err != nil {
		sb.logger.Error("🚀 Error getting system.", "error", err)
		sbCh <- *sb
		return
	}

	filteredWaypoints := lib.Filter(*waypoints, func(waypoint m.Waypoint) bool {
		return waypoint.Type == waypointType
	})

	currentWaypoint := lib.Filter(*waypoints, func(waypoint m.Waypoint) bool {
		return waypoint.Symbol == sb.ship.Nav.WaypointSymbol
	})[0]

	nearestWaypoint, err := lib.NearestWaypoint(&currentWaypoint, &filteredWaypoints)
	if err != nil {
		sb.logger.Error("🚀 Error getting nearest waypoint.", "error", err)
		sbCh <- *sb
		return
	}

	// Navigate to waypoint.
	sb.logger.Infof("🚀 Navigating to nearest %s...", waypointType)

	res, err := sb.client.NavigateShip(sb.ship.Symbol, nearestWaypoint.Symbol)
	if err != nil {
		sb.logger.Error("🚀 Error navigating to waypoint.", "error", err)
		sbCh <- *sb
		return
	}

	sb.logger.Info("🚀 Navigation successful! Waiting until arrival...", "eta", res.Nav.Route.Arrival)
	sb.ship.Fuel = res.Fuel
	sb.ship.Nav = res.Nav

	// Wait until arrival.
	sb.WaitUntilArrival()

	// Send sb to sbCh.
	sbCh <- *sb
}

// NavigateToNearestWaypointWithTrait: Navigate to nearest waypoint with trait.
func (sb *ShipBot) NavigateToNearestWaypointWithTrait(trait string, sbCh chan ShipBot) {
	sb.logger.Info("Navigating to nearest waypoint with trait...", "trait", trait)

	// Get nearest waypoint with trait.
	waypoints, err := sb.client.ListWaypoints(sb.ship.Nav.SystemSymbol)
	if err != nil {
		sb.logger.Error("🚀 Error getting waypoints.", "error", err)
	}

	filteredWaypoints := lib.Filter(*waypoints, func(waypoint m.Waypoint) bool {
		for _, waypointTrait := range waypoint.Traits {
			if waypointTrait.Symbol == trait {
				return true
			}
		}

		return false
	})

	currentWaypoint := lib.Filter(*waypoints, func(waypoint m.Waypoint) bool {
		return waypoint.Symbol == sb.ship.Nav.WaypointSymbol
	})[0]

	nearestWaypoint, err := lib.NearestWaypoint(&currentWaypoint, &filteredWaypoints)
	if err != nil {
		sb.logger.Error("🚀 Error getting nearest waypoint.", "error", err)
		sbCh <- *sb
		return
	}

	// Navigate to waypoint.
	sb.logger.Infof("🚀 Navigating to nearest waypoint with %s...", trait)

	res, err := sb.client.NavigateShip(sb.ship.Symbol, nearestWaypoint.Symbol)
	if err != nil {
		sb.logger.Error("🚀 Error navigating to waypoint.", "error", err)
		sbCh <- *sb
		return
	}

	sb.logger.Info("🚀 Navigation successful! Waiting until arrival...", "eta", res.Nav.Route.Arrival)
	sb.ship.Fuel = res.Fuel
	sb.ship.Nav = res.Nav

	// Wait until arrival.
	sb.WaitUntilArrival()

	// Send sb to sbCh.
	sbCh <- *sb
}

// NewShipBot creates a new instance of ShipBot.
func NewShipBot(client *api.Client, ship *m.Ship, agent *m.Agent) *ShipBot {
	return &ShipBot{
		client: client,
		logger: log.NewWithOptions(os.Stderr, log.Options{
			ReportTimestamp: true,
			Prefix:          fmt.Sprintf("🚀 %s", ship.Symbol),
		}),
		ship:  ship,
		agent: agent,
	}
}

// DockShip: Dock ship at waypoint.
func (sb *ShipBot) DockShip(sbCh chan ShipBot) {
	sb.logger.Info("Docking ship...")
	nav, err := sb.client.DockShip(sb.ship.Symbol)
	if err != nil {
		sb.logger.Error("🚀 Error docking ship.", "error", err)
	}

	sb.ship.Nav = *nav

	sbCh <- *sb
}

// WaitUntilArrival: Wait until ship arrives at its destination.
func (sb *ShipBot) WaitUntilArrival() {
	if sb.ship.Nav.Route.Arrival.Before(time.Now()) {
		sb.logger.Info("Not in transit. Skipping wait.")
		return
	}

	sb.logger.Info("In transit. Waiting until arrival...", "arrival", sb.ship.Nav.Route.Arrival)
	time.Sleep(time.Until(sb.ship.Nav.Route.Arrival))
}

// WaitUntilCooldown: Wait until ship's cooldown expires.
func (sb *ShipBot) WaitUntilCooldown() {
	sb.logger.Info("Waiting until cooldown expires...", "cooldown", sb.cooldown.Expiration)
	if sb.cooldown.Expiration.Before(time.Now()) {
		sb.logger.Info("⚛ Reactor ready. Skipping wait.")
		return
	}

	sb.logger.Info("⚛ Reactor cooldown active. Waiting...", "cooldown", sb.cooldown.Expiration)
	time.Sleep(time.Until(sb.cooldown.Expiration))
}

// IsFullOfCargo checks if the ship is full of cargo, returning a boolean.
func (sb *ShipBot) IsFullOfCargo() bool {
	return sb.ship.Cargo.Units >= sb.ship.Cargo.Capacity
}

// IsAtWaypointOfType checks if the ship is at a waypoint of a given type, returning a boolean.
func (sb *ShipBot) IsAtWaypointOfType(waypointType string) bool {
	waypoint, err := sb.client.GetWaypoint(sb.ship.Nav.SystemSymbol, sb.ship.Nav.WaypointSymbol)
	if err != nil {
		sb.logger.Error("Error getting waypoint.", "error", err)
	}

	return waypoint.Type == waypointType
}

// IsAtWaypointWithTrait checks if the ship is at a waypoint with a given trait, returning a boolean.
func (sb *ShipBot) IsAtWaypointWithTrait(traitSymbol string) bool {
	waypoint, err := sb.client.GetWaypoint(sb.ship.Nav.SystemSymbol, sb.ship.Nav.WaypointSymbol)
	if err != nil {
		sb.logger.Error("Error getting waypoint.", "error", err)
	}

	matchingTraits := lib.Filter(waypoint.Traits, func(t m.WaypointTrait) bool {
		return t.Symbol == traitSymbol
	})

	return len(matchingTraits) > 0
}

// HasStatus checks if the ship has a given status, returning a boolean.
func (sb *ShipBot) HasStatus(status string) bool {
	return sb.ship.Nav.Status == status
}

func (sb *ShipBot) SellCargo(sbCh chan ShipBot) {
	for {
		if sb.ship.Cargo.Units > 0 {
			for _, good := range sb.ship.Cargo.Inventory {
				if lib.Contains(sb.priorities, good.Symbol) {
					sb.logger.Info("💲 Selling priority cargo...", "type", good.Symbol, "units", good.Units)
					res, err := sb.client.SellCargo(sb.ship.Symbol, good.Symbol, good.Units)
					if err != nil {
						sb.logger.Error("💲 Error selling cargo.", "error", err)
						break
					}

					sb.logger.Info("💲 Cargo sold.", "type", res.Transaction.TradeSymbol, "units", res.Transaction.Units, "unitPrice", res.Transaction.PricePerUnit, "totalPrice", res.Transaction.TotalPrice)

					sb.logger.Info("📦 Cargo status updated.", "cargoStatus", fmt.Sprintf("%d/%d", res.Cargo.Units, res.Cargo.Capacity))
					sb.ship.Cargo = res.Cargo

					sb.logger.Info("💰 Agent credits updated.", "credits", res.Agent.Credits)
					sb.agent.Credits = res.Agent.Credits
				} else {
					sb.logger.Info("💲 Selling non-priority cargo...", "type", good.Symbol, "units", good.Units)
					res, err := sb.client.SellCargo(sb.ship.Symbol, good.Symbol, good.Units)
					if err != nil {
						sb.logger.Error("💲 Error selling cargo. Returning to agent...", "error", err)
						break
					}

					sb.logger.Info("💲 Cargo sold.", "type", res.Transaction.TradeSymbol, "units", res.Transaction.Units, "unitPrice", res.Transaction.PricePerUnit, "totalPrice", res.Transaction.TotalPrice)

					sb.logger.Info("📦 Cargo status updated.", "cargoStatus", fmt.Sprintf("%d/%d", res.Cargo.Units, res.Cargo.Capacity))
					sb.ship.Cargo = res.Cargo

					sb.logger.Info("💰 Agent credits updated.", "credits", res.Agent.Credits)
					sb.agent.Credits = res.Agent.Credits
				}
			}
		} else {
			sb.logger.Info("📦 Cargo empty. Reporting to agent...")
			break
		}
	}

	sbCh <- *sb
}

func (sb *ShipBot) ExtractResources(sbCh chan ShipBot) {
	for {
		if !sb.IsFullOfCargo() {
			sb.WaitUntilCooldown()

			res, err := sb.client.ExtractResources(sb.ship.Symbol)
			if err != nil {
				sb.logger.Error(err)
				sb.logger.Info("Mission failed. Reporting to agent...")
				break
			}
			sb.logger.Info("⛏ Resources extracted.", "type", res.Extraction.Yield.Symbol, "units", res.Extraction.Yield.Units)

			// Update cargo
			sb.ship.Cargo = res.Cargo
			sb.logger.Info("📦 Cargo status updated.", "cargoStatus", fmt.Sprintf("%d/%d", res.Cargo.Units, res.Cargo.Capacity))

			// Update cooldown
			sb.cooldown = &res.Cooldown
		} else {
			sb.logger.Info("📦 Cargo full. Reporting to agent...")
			break
		}
	}

	sbCh <- *sb
}

func (sb *ShipBot) GetShipCooldown() (*m.Cooldown, error) {
	cooldown, err := sb.client.GetShipCooldown(sb.ship.Symbol)
	if err != nil {
		return nil, err
	}

	return cooldown, nil
}

// InitiateRequisitionProtocol sends a command ship to search for shipyard and purchase more ships.
func (sb *ShipBot) InitiateRequisitionProtocol(wg *sync.WaitGroup) {
	defer wg.Done()

	sb.logger.Info("Initiating requisition protocol...")

	// Find shipyards in current system
	sb.logger.Info("🔎 Finding shipyards in current system...", "system", sb.ship.Nav.SystemSymbol)
	waypoints, err := sb.FindWaypointsByTrait(sb.ship.Nav.SystemSymbol, "SHIPYARD")
	if err != nil {
		sb.logger.Error("🔎 Error finding shipyards in current system.", "error", err)
	}

	if len(*waypoints) == 0 {
		panic("no shipyards found. Not yet handled.")
	}

	sb.logger.Info("🔎 Shipyards found.", "count", len(*waypoints))

	for _, waypoint := range *waypoints {
		// Travel to shipyard
		sb.logger.Info("🚀 Traveling to shipyard...", "waypoint", waypoint.Symbol)
		sb.NavigateShip(waypoint.Symbol)
	}
}

// NavigateShip sends a ship to a waypoint.
func (sb *ShipBot) NavigateShip(waypointSymbol string) {
	// Check if ship is already at waypoint
	if sb.ship.Nav.WaypointSymbol == waypointSymbol && sb.ship.Nav.Route.Arrival.Before(time.Now()) {
		sb.logger.Info("🚀 Already at waypoint. Navigation skipped.", "waypoint", waypointSymbol)
		return
	}

	// Check if ship is already traveling to waypoint
	if sb.ship.Nav.Route.Arrival.After(time.Now()) {
		if sb.ship.Nav.Route.Destination.Symbol == waypointSymbol {
			sb.logger.Info("🚀 Already traveling to waypoint. Navigation skipped.", "waypoint", waypointSymbol)
			sb.WaitUntilArrival()
			return
		}

		sb.WaitUntilArrival()
	}

	_, err := sb.client.NavigateShip(sb.ship.Symbol, waypointSymbol)
	if err != nil {
		sb.logger.Error("🚀 Error traveling to shipyard.", "error", err)
	}
}

func (sb *ShipBot) FindWaypointsByTrait(systemSymbol, trait string) (*[]m.Waypoint, error) {
	waypoints, err := sb.client.ListWaypoints(systemSymbol)
	if err != nil {
		return nil, err
	}

	waypointsWithTrait := lib.Filter(*waypoints, func(w m.Waypoint) bool {
		for _, t := range w.Traits {
			if t.Symbol == trait {
				return true
			}
		}

		return false
	})

	return &waypointsWithTrait, nil
}
