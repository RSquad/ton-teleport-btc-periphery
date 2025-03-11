package helpers

import "github.com/rsquad/ton-teleport-btc-periphery/frost"

// Helpers

func ConvertMapToFrostPackages(origMap map[string][]byte) (frostMap map[frost.Identifier]frost.Package) {
	frostMap = make(map[frost.Identifier]frost.Package)
	for k, v := range origMap {
		id, _ := frost.DecodeIdentifier(k)
		frostMap[*id] = frost.NewPackage(v)
	}
	return
}
