package pegoutsigner

type CommitmentPackage struct {
	Identifier string
	Package    []byte
}

type SigningShare struct {
	Identifier string
	Package    []byte
	Index      string
}
