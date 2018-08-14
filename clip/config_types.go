package clip

type configFull struct {
	Config        configLong             `json:"config"`
	Lights        map[string]interface{} `json:"lights"`
	Groups        map[string]interface{} `json:"groups"`
	Schedules     map[string]interface{} `json:"schedules"`
	Scenes        map[string]interface{} `json:"scenes"`
	Rules         map[string]interface{} `json:"rules"`
	Sensors       map[string]interface{} `json:"sensors"`
	ResourceLinks map[string]interface{} `json:"resourcelinks"`
}

type configShort struct {
	Name             *string `json:"name"`
	SwVersion        *string `json:"swversion"`
	APIVersion       *string `json:"apiversion"`
	Mac              *string `json:"mac"`
	BridgeID         *string `json:"bridgeid"`
	ReplacesBridgeID *string `json:"replacesbridgeid"`
	FactoryNew       bool    `json:"factorynew"`
	DatastoreVersion string  `json:"datastoreversion"`
	StarterKitID     string  `json:"starterkitid"`
}

type configLong struct {
	*configShort
	ZigbeeChannel    int                       `json:"zigbeechannel"`
	Dhcp             bool                      `json:"dhcp"`
	IPAddress        string                    `json:"ipaddress"`
	Netmask          string                    `json:"netmask"`
	Gateway          string                    `json:"gateway"`
	ProxyAddress     string                    `json:"proxyaddress"`
	ProxyPort        int                       `json:"proxyport"`
	UTC              string                    `json:"UTC"`
	LocalTime        string                    `json:"localtime"`
	TimeZone         string                    `json:"timezone"`
	SwUpdate         swUpdate                  `json:"swupdate"`
	SwUpdate2        swUpdateNew               `json:"swupdate2"`
	LinkButton       bool                      `json:"linkbutton"`
	PortalServices   bool                      `json:"portalservices"`
	PortalConnection *string                   `json:"portalconnection"`
	PortalState      configPortalState         `json:"portalstate"`
	InternetServices configInternetServices    `json:"internetservices"`
	Backup           configBackup              `json:"backup"`
	Whitelist        map[string]whitelistEntry `json:"whitelist"`
}

type configPortalState struct {
	SignedOn      bool   `json:"signedon"`
	Incoming      bool   `json:"incoming"`
	Outgoing      bool   `json:"outgoing"`
	Communication string `json:"communication"`
}

type configInternetServices struct {
	Internet     string `json:"internet"`
	RemoteAccess string `json:"remoteaccess"`
	Time         string `json:"time"`
	SwUpdate     string `json:"swupdate"`
}

type configBackup struct {
	Status    string `json:"status"`
	ErrorCode int    `json:"errorcode"`
}

type whitelistEntry struct {
	LastUseDate string `json:"last use date"`
	CreateDate  string `json:"create date"`
	Name        string `json:"name"`
}

type swUpdate struct {
	UpdateState    int                 `json:"updatestate"`
	CheckForUpdate bool                `json:"checkforupdate"`
	DeviceTypes    swUpdateDeviceTypes `json:"devicetypes"`
}
type swUpdateNew struct {
	CheckForUpdate bool                `json:"checkforupdate"`
	LastChange     string              `json:"lastchange"`
	Bridge         swUpdateBridge      `json:"bridge"`
	State          string              `json:"state"`
	AutoInstall    swUpdateAutoInstall `json:"autoinstall"`
}

type swUpdateBridge struct {
	State       string `json:"state"`
	LastInstall string `json:"lastinstall"`
}

type swUpdateAutoInstall struct {
	UpdateTime string `json:"updatetime"`
	On         bool   `json:"on"`
}

type swUpdateDeviceTypes struct {
	Bridge  bool          `json:"bridge"`
	Lights  []interface{} `json:"lights"`
	Sensors []interface{} `json:"sensors"`
	URL     string        `json:"url"`
	Text    string        `json:"text"`
	Notify  bool          `json:"notify"`
}
