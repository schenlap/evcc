package wrapper

import (
	"math"
	"time"

	"github.com/andig/evcc/api"
	"github.com/andig/evcc/util"
)

// SocEstimator provides vehicle soc and charge duration
// Vehicle SoC can be estimated to provide more granularity
type SocEstimator struct {
	log      *util.Logger
	vehicle  api.Vehicle
	estimate bool

	capacity          float64 // vehicle capacity in Wh cached to simplify testing
	socCharge         float64 // estimated vehicle SoC
	prevSoC           float64 // previous vehicle SoC in %
	prevChargedEnergy float64 // previous charged energy in Wh
	energyPerSocStep  float64 // Energy per SoC percent in Wh
}

// NewSocEstimator creates new estimator
func NewSocEstimator(log *util.Logger, vehicle api.Vehicle, estimate bool) *SocEstimator {
	s := &SocEstimator{
		log:      log,
		vehicle:  vehicle,
		estimate: estimate,
	}

	s.Reset()

	return s
}

// Reset resets the estimation process to default values
func (s *SocEstimator) Reset() {
	s.prevSoC = 0
	s.prevChargedEnergy = 0
	s.capacity = float64(s.vehicle.Capacity()) * 1e3 // cache to simplify debugging
	s.energyPerSocStep = s.capacity / 100
}

// RemainingChargeDuration returns the remaining duration estimate based on SoC, target and charge power
func (s *SocEstimator) RemainingChargeDuration(chargePower float64, targetSoC int) time.Duration {
	if chargePower > 0 {
		percentRemaining := float64(targetSoC) - s.socCharge
		if percentRemaining <= 0 {
			return 0
		}

		whRemaining := percentRemaining / 100 * s.capacity
		return time.Duration(float64(time.Hour) * whRemaining / chargePower).Round(time.Second)
	}

	return -1
}

// SoC implements Vehicle.ChargeState with addition of given charged energy
func (s *SocEstimator) SoC(chargedEnergy float64) (float64, error) {
	f, err := s.vehicle.ChargeState()
	if err != nil {
		return s.socCharge, err
	}

	s.socCharge = f

	if s.estimate {
		socDelta := s.socCharge - s.prevSoC
		energyDelta := math.Max(chargedEnergy, 0) - s.prevChargedEnergy

		if socDelta != 0 || energyDelta < 0 { // soc value change or unexpected energy reset
			// calculate gradient, wh per soc %
			if socDelta > 1 && energyDelta > 0 && s.prevSoC > 0 {
				s.energyPerSocStep = energyDelta / socDelta
				s.log.TRACE.Printf("soc gradient updated: energyPerSocStep: %0.0fWh, virtualBatCap: %0.1fkWh", s.energyPerSocStep, s.energyPerSocStep*100/1e3)
			}

			// sample charged energy at soc change, reset energy delta
			s.prevChargedEnergy = math.Max(chargedEnergy, 0)
			s.prevSoC = s.socCharge
		} else {
			s.socCharge = math.Min(f + energyDelta / s.energyPerSocStep, 100)
			s.log.TRACE.Printf("soc estimated: %.2f%% (vehicle: %.2f%%)", s.socCharge, f)
		}
	}

	return s.socCharge, nil
}
