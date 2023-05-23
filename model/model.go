package model

import "time"

type Agent struct {
	AccountId    string `json:"accountId"`
	Symbol       string `json:"symbol"`
	Headquarters string `json:"headquarters"`
	Credits      int    `json:"credits"`
}

type Chart struct {
	WaypointSymbol string    `json:"waypointSymbol"`
	SubmittedBy    string    `json:"submittedBy"`
	SubmittedOn    time.Time `json:"submittedOn"`
}

type ConnectedSystem struct {
	Symbol        string `json:"symbol"`
	SectorSymbol  string `json:"sectorSymbol"`
	Type          string `json:"type"`
	FactionSymbol string `json:"factionSymbol"`
	X             int    `json:"x"`
	Y             int    `json:"y"`
	Distance      int    `json:"distance"`
}

type Contract struct {
	ID            string        `json:"id"`
	FactionSymbol string        `json:"factionSymbol"`
	Type          string        `json:"type"`
	Terms         ContractTerms `json:"terms"`
	Accepted      bool          `json:"accepted"`
	Fulfilled     bool          `json:"fulfilled"`
	Expiration    time.Time     `json:"expiration"`
}

type ContractDeliverGood struct {
	TradeSymbol       string `json:"tradeSymbol"`
	DestinationSymbol string `json:"destinationSymbol"`
	UnitsRequired     int    `json:"unitsRequired"`
	UnitsFulfilled    int    `json:"unitsFulfilled"`
}

type ContractPayment struct {
	OnAccepted  int `json:"onAccepted"`
	OnFulfilled int `json:"onFulfilled"`
}

type ContractTerms struct {
	Deadline time.Time             `json:"deadline"`
	Payment  ContractPayment       `json:"payment"`
	Deliver  []ContractDeliverGood `json:"deliver"`
}

type Cooldown struct {
	ShipSymbol       string    `json:"shipSymbol"`
	TotalSeconds     int       `json:"totalSeconds"`
	RemainingSeconds int       `json:"remainingSeconds"`
	Expiration       time.Time `json:"expiration"`
}

type Extraction struct {
	ShipSymbol string `json:"shipSymbol"`
	Yield      struct {
		Symbol string `json:"symbol"`
		Units  int    `json:"units"`
	} `json:"yield"`
}

type JumpGate struct {
	JumpRange        int               `json:"jumpRange"`
	FactionSymbol    string            `json:"factionSymbol"`
	ConnectedSystems []ConnectedSystem `json:"connectedSystems"`
}

type Market struct {
	Symbol       string              `json:"symbol"`
	Exports      []TradeGood         `json:"exports"`
	Imports      []TradeGood         `json:"imports"`
	Exchange     []TradeGood         `json:"exchange"`
	Transactions []MarketTransaction `json:"transactions"`
	TradeGoods   []MarketTradeGood   `json:"tradeGoods"`
}

type MarketTradeGood struct {
	Symbol        string `json:"symbol"`
	TradeVolume   int    `json:"tradeVolume"`
	Supply        string `json:"supply"`
	PurchasePrice int    `json:"purchasePrice"`
	SellPrice     int    `json:"sellPrice"`
}

type MarketTransaction struct {
	WaypointSymbol string    `json:"waypointSymbol"`
	ShipSymbol     string    `json:"shipSymbol"`
	TradeSymbol    string    `json:"tradeSymbol"`
	Type           string    `json:"type"`
	Units          int       `json:"units"`
	PricePerUnit   int       `json:"pricePerUnit"`
	TotalPrice     int       `json:"totalPrice"`
	Timestamp      time.Time `json:"timestamp"`
}

type Ship struct {
	Symbol       string           `json:"symbol"`
	Registration ShipRegistration `json:"registration"`
	Nav          ShipNav          `json:"nav"`
	Crew         ShipCrew         `json:"crew"`
	Frame        ShipFrame        `json:"frame"`
	Reactor      ShipReactor      `json:"reactor"`
	Engine       ShipEngine       `json:"engine"`
	Modules      []ShipModule     `json:"modules"`
	Mounts       []ShipMount      `json:"mounts"`
	Cargo        ShipCargo        `json:"cargo"`
	Fuel         ShipFuel         `json:"fuel"`
}

type ShipCargo struct {
	Capacity  int             `json:"capacity"`
	Units     int             `json:"units"`
	Inventory []ShipCargoItem `json:"inventory"`
}

type ShipCargoItem struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Units       int    `json:"units"`
}

type ShipCrew struct {
	Current  int    `json:"current"`
	Required int    `json:"required"`
	Capacity int    `json:"capacity"`
	Rotation string `json:"rotation"`
	Morale   int    `json:"morale"`
	Wages    int    `json:"wages"`
}

type ShipEngine struct {
	Symbol       string           `json:"symbol"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Condition    int              `json:"condition"`
	Speed        int              `json:"speed"`
	Requirements ShipRequirements `json:"requirements"`
}

type ShipFuel struct {
	Current  int `json:"current"`
	Capacity int `json:"capacity"`
	Consumed struct {
		Amount    int       `json:"amount"`
		Timestamp time.Time `json:"timestamp"`
	} `json:"consumed"`
}

type ShipFrame struct {
	Symbol         string           `json:"symbol"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	Condition      int              `json:"condition"`
	ModuleSlots    int              `json:"moduleSlots"`
	MountingPoints int              `json:"mountingPoints"`
	FuelCapacity   int              `json:"fuelCapacity"`
	Requirements   ShipRequirements `json:"requirements"`
}

type ShipMount struct {
	Symbol       string           `json:"symbol"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Strength     int              `json:"strength"`
	Deposits     []string         `json:"deposits"`
	Requirements ShipRequirements `json:"requirements"`
}

type ShipModule struct {
	Symbol       string           `json:"symbol"`
	Capacity     int              `json:"capacity"`
	Range        int              `json:"range"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Requirements ShipRequirements `json:"requirements"`
}

type ShipNav struct {
	SystemSymbol   string       `json:"systemSymbol"`
	WaypointSymbol string       `json:"waypointSymbol"`
	Route          ShipNavRoute `json:"route"`
	Status         string       `json:"status"`
	FlightMode     string       `json:"flightMode"`
}

type ShipNavRoute struct {
	Destination   ShipNavRouteWaypoint `json:"destination"`
	Departure     ShipNavRouteWaypoint `json:"departure"`
	DepartureTime time.Time            `json:"departureTime"`
	Arrival       time.Time            `json:"arrival"`
}

type ShipNavRouteWaypoint struct {
	Symbol       string `json:"symbol"`
	Type         string `json:"type"`
	SystemSymbol string `json:"systemSymbol"`
	X            int    `json:"x"`
	Y            int    `json:"y"`
}

type ShipReactor struct {
	Symbol       string           `json:"symbol"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Condition    int              `json:"condition"`
	PowerOutput  int              `json:"powerOutput"`
	Requirements ShipRequirements `json:"requirements"`
}

type ShipRegistration struct {
	Name          string `json:"name"`
	FactionSymbol string `json:"factionSymbol"`
	Role          string `json:"role"`
}

type ShipRequirements struct {
	Power int `json:"power"`
	Crew  int `json:"crew"`
	Slots int `json:"slots"`
}

type ShipType struct {
	Type string `json:"type"`
}

type Shipyard struct {
	Symbol       string                `json:"symbol"`
	ShipTypes    []ShipType            `json:"shipTypes"`
	Transactions []ShipyardTransaction `json:"transactions"`
	Ships        []ShipyardShip        `json:"ships"`
}

type ShipyardShip struct {
	Type          string       `json:"type"`
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	PurchasePrice int          `json:"purchasePrice"`
	Frame         ShipFrame    `json:"frame"`
	Reactor       ShipReactor  `json:"reactor"`
	Engine        ShipEngine   `json:"engine"`
	Modules       []ShipModule `json:"modules"`
	Mounts        []ShipMount  `json:"mounts"`
}

type ShipyardTransaction struct {
	WaypointSymbol string    `json:"waypointSymbol"`
	ShipSymbol     string    `json:"shipSymbol"`
	Price          int       `json:"price"`
	AgentSymbol    string    `json:"agentSymbol"`
	Timestamp      time.Time `json:"timestamp"`
}

type Survey struct {
	Signature string `json:"signature"`
	Symbol    string `json:"symbol"`
	Deposits  []struct {
		Symbol string `json:"symbol"`
	} `json:"deposits"`
	Expiration time.Time `json:"expiration"`
	Size       string    `json:"size"`
}

type System struct {
	Symbol       string           `json:"symbol"`
	SectorSymbol string           `json:"sectorSymbol"`
	Type         string           `json:"type"`
	X            int              `json:"x"`
	Y            int              `json:"y"`
	Waypoints    []SystemWaypoint `json:"waypoints"`
	Factions     []SystemFaction  `json:"factions"`
}

type SystemFaction struct {
	Symbol string `json:"symbol"`
}

type SystemWaypoint struct {
	Symbol string `json:"symbol"`
	Type   string `json:"type"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}

type Waypoint struct {
	Symbol       string             `json:"symbol"`
	Type         string             `json:"type"`
	SystemSymbol string             `json:"systemSymbol"`
	X            int                `json:"x"`
	Y            int                `json:"y"`
	Orbitals     []WaypointOribital `json:"orbitals"`
	Faction      WaypointFaction    `json:"faction"`
	Traits       []WaypointTrait    `json:"traits"`
	Chart        Chart              `json:"chart"`
}

type WaypointFaction struct {
	Symbol string `json:"symbol"`
}

type WaypointOribital struct {
	Symbol string `json:"symbol"`
}

type WaypointTrait struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TradeGood struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
