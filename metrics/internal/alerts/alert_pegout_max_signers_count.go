package alerts

import "math/big"

type AlertPegoutMaxSignersCount struct {
}

func NewAlertPegoutMaxSignersCount() Alert {
	return &AlertPegoutMaxSignersCount{}
}

func (alert *AlertPegoutMaxSignersCount) Check(dataSource AlertDataSource) (int, error) {
	configuratorContractData, err := dataSource.ConfiguratorContractData()
	if err != nil {
		return -1, err
	}

	prevDkg, err := dataSource.PrevDkg()
	if err != nil {
		return -1, err
	}

	unsignedPegout, err := dataSource.FirstUnsignedPegout()
	if err != nil {
		return -1, err
	}

	if unsignedPegout == nil {
		return -1, nil
	}

	//unsignedPegout.ExpiredAt
	mask := new(big.Int).Or(
		unsignedPegout.CommitmentsMaskAccepted,
		unsignedPegout.CommitmentsMaskOther,
	)

	if (mask < ) ?

	return -1, nil
}
