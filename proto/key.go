package proto

func NewVendorApp(vendorId, appId string) *VendorApp {
	return &VendorApp{
		VendorId: vendorId,
		AppId:    appId,
	}
}
