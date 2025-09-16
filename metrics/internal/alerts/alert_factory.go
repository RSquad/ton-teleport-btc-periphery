package alerts

import "github.com/xssnick/tonutils-go/address"

type AlertFactoryFn func() Alert

type AlertFactory struct {
	Factories map[string]AlertFactoryFn
}

func NewAlertFactory(contractAddrs map[string]*address.Address) *AlertFactory {
	factories := make(map[string]AlertFactoryFn)

	// alert_pegout_fee_not_reset (pegout.fee.not.reset)
	factories["alert_pegout_fee_not_reset"] = func() Alert {
		return NewAlertPegoutFeeNotReset()
	}

	// alert_pegout_signers (pegout.signers)
	factories["alert_pegout_signers"] = func() Alert {
		return NewAlertPegoutSigners()
	}

	// alert_pegout_restarts (pegout.restarts)
	factories["alert_pegout_restarts"] = func() Alert {
		return NewAlertPegoutRestarts()
	}

	// alert_pegout_commintments (pegout.commitments)
	factories["alert_pegout_commintments"] = func() Alert {
		return NewAlertPegoutCommintments()
	}

	// alert_pegout_in_mempool (pegout.in.mempool)
	factories["alert_pegout_in_mempool"] = func() Alert {
		return NewAlertPegoutInMempool()
	}

	// alert_pegout_signing_duration (pegout.signing.duration)
	factories["alert_pegout_signing_duration"] = func() Alert {
		return NewAlertPegoutSigningDuration()
	}

	// dkg_status (dkg.status)
	factories["dkg_status"] = func() Alert {
		return NewAlertDkgStatus()
	}

	// alert_total_service_fee (total.service.fee)
	factories["alert_total_service_fee"] = func() Alert {
		return NewAlertTotalServiceFee()
	}

	// alert_dkg_restarts (dkg.restarts)
	factories["alert_dkg_restarts"] = func() Alert {
		return NewAlertDkgRestarts()
	}

	// alert_dkg_participants (dkg.participants)
	factories["alert_dkg_participants"] = func() Alert {
		return NewAlertDkgParticipants()
	}

	// alert_dkg_culprit_found (dkg.culprit.found)
	factories["alert_dkg_culprit_found"] = func() Alert {
		return NewAlertDkgCulpritFound()
	}

	// alert_contract_balance_*
	for name, addr := range contractAddrs {
		factories["alert_contract_balance_"+name] = func() Alert {
			return NewAlertContractBalance("alert_contract_balance_"+name, addr)
		}
	}

	return &AlertFactory{
		Factories: factories,
	}
}
