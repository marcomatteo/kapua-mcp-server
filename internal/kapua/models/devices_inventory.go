package models

// Device inventory models (per specs: deviceInventory, deviceInventoryBundles, etc.)

// InventoryItem represents a single inventory entry returned by Kapua.
type InventoryItem struct {
	Name     string `json:"name,omitempty"`
	Version  string `json:"version,omitempty"`
	ItemType string `json:"itemType,omitempty"`
}

// DeviceInventory wraps the list of inventory items installed on a device.
type DeviceInventory struct {
	InventoryItems []InventoryItem `json:"inventoryItems,omitempty"`
}

// DeviceInventoryBundle models a single OSGi bundle entry from inventory.
type DeviceInventoryBundle struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	Status  string `json:"status,omitempty"`
	Signed  bool   `json:"signed,omitempty"`
}

// DeviceInventoryBundles groups bundle inventory information.
type DeviceInventoryBundles struct {
	InventoryBundles []DeviceInventoryBundle `json:"inventoryBundles,omitempty"`
}

// DeviceInventoryContainerState enumerates possible inventory container states.
type DeviceInventoryContainerState string

const (
	DeviceInventoryContainerStateActive      DeviceInventoryContainerState = "ACTIVE"
	DeviceInventoryContainerStateInstalled   DeviceInventoryContainerState = "INSTALLED"
	DeviceInventoryContainerStateUninstalled DeviceInventoryContainerState = "UNINSTALLED"
	DeviceInventoryContainerStateUnknown     DeviceInventoryContainerState = "UNKNOWN"
)

// DeviceInventoryContainer defines a container instance reported by Kapua.
type DeviceInventoryContainer struct {
	Name          string                        `json:"name,omitempty"`
	Version       string                        `json:"version,omitempty"`
	ContainerType string                        `json:"containerType,omitempty"`
	State         DeviceInventoryContainerState `json:"state,omitempty"`
}

// DeviceInventoryContainers groups container inventory responses.
type DeviceInventoryContainers struct {
	InventoryContainers []DeviceInventoryContainer `json:"inventoryContainers,omitempty"`
}

// DeviceInventoryPackage represents a single package entry (system inventory).
type DeviceInventoryPackage struct {
	Name        string `json:"name,omitempty"`
	Version     string `json:"version,omitempty"`
	PackageType string `json:"packageType,omitempty"`
}

// DeviceInventoryPackages groups system packages returned by Kapua.
type DeviceInventoryPackages struct {
	SystemPackages []DeviceInventoryPackage `json:"systemPackages,omitempty"`
}

// DeviceInventoryDeploymentPackage captures deployment package metadata.
type DeviceInventoryDeploymentPackage struct {
	Name           string                  `json:"name,omitempty"`
	Version        string                  `json:"version,omitempty"`
	PackageBundles []DeviceInventoryBundle `json:"packageBundles,omitempty"`
}

// DeviceInventoryDeploymentPackages groups deployment packages inventory data.
type DeviceInventoryDeploymentPackages struct {
	DeploymentPackages []DeviceInventoryDeploymentPackage `json:"deploymentPackages,omitempty"`
	SystemPackages     []DeviceInventoryDeploymentPackage `json:"systemPackages,omitempty"`
}
