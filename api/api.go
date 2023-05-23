package api

import (
	"errors"
	"net/url"
	"sync"
	"time"

	m "github.com/GeoffreyDick/gogarin/model"
	resty "github.com/go-resty/resty/v2"
)

/*
🐌 Throttle
*/
type Throttle struct {
	MaxRequestsPerSecond int
	LastRequestTime      time.Time
	Mutex                sync.Mutex
}

func NewThrottle(maxRequestsPerSecond int) *Throttle {
	return &Throttle{
		MaxRequestsPerSecond: maxRequestsPerSecond,
		LastRequestTime:      time.Now(),
	}
}

func (t *Throttle) Wait() {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()

	timeSinceLastRequest := time.Since(t.LastRequestTime)

	timeToWait := time.Duration(float64(time.Second) / float64(t.MaxRequestsPerSecond))

	if timeSinceLastRequest < timeToWait {
		time.Sleep(timeToWait - timeSinceLastRequest)
	}

	t.LastRequestTime = time.Now()
}

/*
💻 Client
*/
type Client struct {
	r *resty.Client
	t *Throttle
}

func NewClient(token string) *Client {
	bearer := "Bearer " + token

	r := resty.
		New().
		SetBaseURL(baseURL.String()).
		SetHeader("Authorization", bearer).
		SetTimeout(1*time.Minute).
		SetHeader("Accept", "application/json").
		EnableTrace()

	t := NewThrottle(2)

	return &Client{r, t}
}

/*
📨 spacetraders.io API
*/
type ErrorResponse struct {
	Error struct {
		Message string      `json:"message"`
		Code    int         `json:"code"`
		Data    interface{} `json:"data"`
	}
}

var baseURL = url.URL{
	Scheme: "https",
	Host:   "api.spacetraders.io",
	Path:   "/v2",
}

func (c *Client) GetMyAgent() (*m.Agent, error) {
	c.t.Wait()

	var resultResponse struct {
		Data m.Agent `json:"data"`
	}

	url := "/my/agent"

	res, err := c.r.R().
		SetResult(&resultResponse).
		SetError(ErrorResponse{}).
		Get(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(*ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

func (c *Client) GetMyContracts() (*[]m.Contract, error) {
	c.t.Wait()

	var resultResponse struct {
		Data []m.Contract `json:"data"`
	}

	url := "/my/contracts"

	res, err := c.r.R().
		SetResult(&resultResponse).
		SetError(ErrorResponse{}).
		Get(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(*ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

type AcceptContractResponse struct {
	Agent    m.Agent    `json:"agent"`
	Contract m.Contract `json:"contract"`
}

// AcceptContract accepts a contract.
func (c *Client) AcceptContract(contractId string) (*AcceptContractResponse, error) {
	c.t.Wait()

	var resultResponse struct {
		Data AcceptContractResponse `json:"data"`
	}

	url := "/my/contracts/" + contractId + "/accept"

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resultResponse).
		SetError(ErrorResponse{}).
		Post(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(*ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

func (c *Client) GetMyShips() (*[]m.Ship, error) {
	c.t.Wait()

	var resultResponse struct {
		Data []m.Ship `json:"data"`
	}

	url := "/my/ships"

	res, err := c.r.R().
		SetResult(&resultResponse).
		SetError(ErrorResponse{}).
		Get(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(*ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

func (c *Client) GetShipCooldown(shipSymbol string) (*m.Cooldown, error) {
	c.t.Wait()

	var resultResponse struct {
		Data m.Cooldown `json:"data"`
	}

	url := "/my/ships/" + shipSymbol + "/cooldown"

	res, err := c.r.R().
		SetResult(&resultResponse).
		Get(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(*ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

type NavigateShipResponse struct {
	Fuel m.ShipFuel `json:"fuel"`
	Nav  m.ShipNav  `json:"nav"`
}

// NavigateShip: Navigate to a target destination. The destination must be located within the same system as the ship. Navigating will consume the necessary fuel and supplies from the ship's manifest, and will pay out crew wages from the agent's account.
//
// The returned response will detail the route information including the expected time of arrival. Most ship actions are unavailable until the ship has arrived at it's destination.
//
// To travel between systems, see the ship's warp or jump actions.
func (c *Client) NavigateShip(shipSymbol string, waypointSymbol string) (*NavigateShipResponse, error) {
	c.t.Wait()

	var resultResponse struct {
		Data NavigateShipResponse `json:"data"`
	}

	url := "/my/ships/" + shipSymbol + "/navigate"

	body := struct {
		WaypointSymbol string `json:"waypointSymbol"`
	}{waypointSymbol}

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(&resultResponse).
		SetError(&ErrorResponse{}).
		Post(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		message := res.Error().(*ErrorResponse).Error.Message
		return nil, errors.New(message)
	}

	return &resultResponse.Data, nil
}

func (c *Client) OrbitShip(shipSymbol string) (*m.ShipNav, error) {
	c.t.Wait()

	var resultResponse struct {
		Data m.ShipNav `json:"data"`
	}

	url := "/my/ships/" + shipSymbol + "/orbit"

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resultResponse).
		Post(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(*ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

func (c *Client) DockShip(shipSymbol string) (*m.ShipNav, error) {
	c.t.Wait()

	var resultResponse struct {
		Data m.ShipNav `json:"data"`
	}

	url := "/my/ships/" + shipSymbol + "/dock"

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resultResponse).
		Post(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(*ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

type CreateSurveyResponse struct {
	Cooldown m.Cooldown `json:"cooldown"`
	Surveys  []m.Survey `json:"surveys"`
}

func (c *Client) CreateSurvey(shipSymbol string) (*CreateSurveyResponse, error) {
	c.t.Wait()

	var resultResponse struct {
		Data CreateSurveyResponse `json:"data"`
	}

	url := "/my/ships/" + shipSymbol + "/survey"

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resultResponse).
		SetError(ErrorResponse{}).
		Post(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(*ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

type ExtractResourcesResponse struct {
	Cooldown   m.Cooldown   `json:"cooldown"`
	Extraction m.Extraction `json:"extraction"`
	Cargo      m.ShipCargo  `json:"cargo"`
}

// Extract resources from the waypoint into your ship. Send an optional survey as the payload to target specific yields.
func (c *Client) ExtractResources(shipSymbol string, surveys ...m.Survey) (*ExtractResourcesResponse, error) {
	c.t.Wait()

	var resultResponse struct {
		Data ExtractResourcesResponse `json:"data"`
	}

	url := "/my/ships/" + shipSymbol + "/extract"

	var body interface{}

	// Include the survey if one was provided
	if len(surveys) > 0 {
		body = struct {
			Survey m.Survey `json:"survey"`
		}{surveys[0]}
	}

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		SetResult(&resultResponse).
		SetError(&ErrorResponse{}).
		Post(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(*ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

// Jettison cargo from your ship's cargo hold.
func (c *Client) JettisonCargo(shipSymbol string, cargoSymbol m.TradeGood, units int) (*m.ShipCargo, error) {
	c.t.Wait()

	var resultResponse struct {
		Data m.ShipCargo `json:"data"`
	}

	url := "/my/ships/" + shipSymbol + "/jettison"

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"symbol": cargoSymbol,
			"units":  units,
		}).
		SetResult(&resultResponse).
		Post(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(*ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

// Jump your ship instantly to a target system. Unlike other forms of navigation, jumping requires a unit of antimatter.
func (c *Client) JumpShip(shipSymbol string, systemSymbol string) (*m.ShipNav, error) {
	c.t.Wait()

	var resultResponse struct {
		Data m.ShipNav `json:"data"`
	}

	url := "/my/ships/" + shipSymbol + "/jump"

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"systemSymbol": systemSymbol,
		}).
		SetResult(&resultResponse).
		Post(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(*ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

type SellCargoResponse struct {
	Agent       m.Agent             `json:"agent"`
	Cargo       m.ShipCargo         `json:"cargo"`
	Transaction m.MarketTransaction `json:"transaction"`
}

func (c *Client) SellCargo(shipSymbol string, cargoSymbol string, units int) (*SellCargoResponse, error) {
	c.t.Wait()

	var resultResponse struct {
		Data SellCargoResponse `json:"data"`
	}

	url := "/my/ships/" + shipSymbol + "/sell"

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"symbol": cargoSymbol,
			"units":  units,
		}).
		SetResult(&resultResponse).
		SetError(ErrorResponse{}).
		Post(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(*ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

/*
🌌 Systems
*/

// ListSystems returns a list of all systems.
func (c *Client) ListSystems() (*[]m.System, error) {
	c.t.Wait()

	var resultResponse struct {
		Data []m.System `json:"data"`
	}

	url := "/systems"

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resultResponse).
		SetError(ErrorResponse{}).
		Get(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

// GetSystem gets the details of a system.
func (c *Client) GetSystem(systemSymbol string) (*m.System, error) {
	c.t.Wait()

	var resultResponse struct {
		Data m.System `json:"data"`
	}

	url := "/systems/" + systemSymbol

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resultResponse).
		SetError(ErrorResponse{}).
		Get(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

// ListWaypoints fetches all of the waypoints for a given system. System must be charted or a ship must be present to return waypoint details.
func (c *Client) ListWaypoints(systemSymbol string) (*[]m.Waypoint, error) {
	c.t.Wait()

	var resultResponse struct {
		Data []m.Waypoint `json:"data"`
	}

	url := "/systems/" + systemSymbol + "/waypoints"

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resultResponse).
		SetError(ErrorResponse{}).
		Get(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

// GetWaypoint views the details of a waypoint.
func (c *Client) GetWaypoint(systemSymbol string, waypointSymbol string) (*m.Waypoint, error) {
	c.t.Wait()

	var resultResponse struct {
		Data m.Waypoint `json:"data"`
	}

	url := "/systems/" + systemSymbol + "/waypoints/" + waypointSymbol

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resultResponse).
		SetError(ErrorResponse{}).
		Get(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(*ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

// GetMarket: Retrieve imports, exports and exchange data from a marketplace. Imports can be sold, exports can be purchased, and exchange goods can be purchased or sold. Send a ship to the waypoint to access trade good prices and recent transactions.
func (c *Client) GetMarket(systemSymbol string, waypointSymbol string) (*m.Market, error) {
	c.t.Wait()

	var resultResponse struct {
		Data m.Market `json:"data"`
	}

	url := "/systems/" + systemSymbol + "/waypoints/" + waypointSymbol + "/market"

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resultResponse).
		SetError(ErrorResponse{}).
		Get(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

// GetShipyard: Get the shipyard for a waypoint. Send a ship to the waypoint to access ships that are currently available for purchase and recent transactions.
func (c *Client) GetShipyard(systemSymbol string, waypointSymbol string) (*m.Shipyard, error) {
	c.t.Wait()

	var resultResponse struct {
		Data m.Shipyard `json:"data"`
	}

	url := "/systems/" + systemSymbol + "/waypoints/" + waypointSymbol + "/shipyard"

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resultResponse).
		SetError(ErrorResponse{}).
		Get(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}

// GetJumpGate: Get jump gate details for a waypoint.
func (c *Client) GetJumpGate(systemSymbol string, waypointSymbol string) (*m.JumpGate, error) {
	c.t.Wait()

	var resultResponse struct {
		Data m.JumpGate `json:"data"`
	}

	url := "/systems/" + systemSymbol + "/waypoints/" + waypointSymbol + "/jumpgate"

	res, err := c.r.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&resultResponse).
		SetError(ErrorResponse{}).
		Get(url)
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, errors.New(res.Error().(ErrorResponse).Error.Message)
	}

	return &resultResponse.Data, nil
}
