package fine

import (
	"testing"
	"time"
)

func at(hour int) time.Time {
	return time.Date(2026, 1, 15, hour, 0, 0, 0, time.UTC)
}

func TestTimeMultiplier_DayNightBoundaries(t *testing.T) {
	tm := DefaultRuleset().TimeMultiplier // day [6,22): 1.0, night: 1.5

	cases := []struct {
		hour int
		want float64
	}{
		{0, 1.5},  // overnight
		{5, 1.5},  // just before day starts
		{6, 1.0},  // day starts (inclusive)
		{12, 1.0}, // midday
		{21, 1.0}, // last day hour
		{22, 1.5}, // night starts (inclusive)
		{23, 1.5}, // late night
	}
	for _, c := range cases {
		if got := tm.For(at(c.hour)); got != c.want {
			t.Errorf("hour %02d:00 => %.1f, want %.1f", c.hour, got, c.want)
		}
	}
}

func TestRepeatMultiplier_Tiers(t *testing.T) {
	rm := DefaultRuleset().RepeatMultiplier // 0:1.0, 1:1.5, >=2:2.0

	cases := []struct {
		prior int
		want  float64
	}{
		{0, 1.0},
		{1, 1.5},
		{2, 2.0},
		{5, 2.0}, // anything >= 2 stays at the top tier
	}
	for _, c := range cases {
		if got := rm.For(c.prior); got != c.want {
			t.Errorf("prior=%d => %.1f, want %.1f", c.prior, got, c.want)
		}
	}
}

func TestCalculate_BaseAmountsPerType(t *testing.T) {
	r := DefaultRuleset()
	want := map[string]int64{
		"expired_meter":    50_000,
		"no_parking_zone":  150_000,
		"blocking_hydrant": 250_000,
		"disabled_spot":    500_000,
	}
	for typ, base := range want {
		// daytime, no priors -> multipliers are 1.0, so final == base.
		b, err := r.Calculate(Input{ViolationType: typ, OccurredAt: at(10), PriorUnpaidCount: 0})
		if err != nil {
			t.Fatalf("%s: %v", typ, err)
		}
		if b.BaseAmount != base || b.FinalAmount != base {
			t.Errorf("%s: base=%d final=%d, want %d", typ, b.BaseAmount, b.FinalAmount, base)
		}
	}
}

func TestCalculate_EndToEnd(t *testing.T) {
	r := DefaultRuleset()
	cases := []struct {
		name      string
		typ       string
		hour      int
		prior     int
		wantTime  float64
		wantRepe  float64
		wantFinal int64
	}{
		{"day, no priors", "expired_meter", 10, 0, 1.0, 1.0, 50_000},
		{"night, no priors", "blocking_hydrant", 23, 0, 1.5, 1.0, 375_000},
		{"day, 1 prior", "no_parking_zone", 9, 1, 1.0, 1.5, 225_000},
		{"night, 2 priors", "disabled_spot", 2, 2, 1.5, 2.0, 1_500_000},
		{"night boundary 22:00, 1 prior", "no_parking_zone", 22, 1, 1.5, 1.5, 337_500},
		{"day boundary 06:00, 3 priors", "expired_meter", 6, 3, 1.0, 2.0, 100_000},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			b, err := r.Calculate(Input{ViolationType: c.typ, OccurredAt: at(c.hour), PriorUnpaidCount: c.prior})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if b.TimeMultiplier != c.wantTime {
				t.Errorf("time mult = %.2f, want %.2f", b.TimeMultiplier, c.wantTime)
			}
			if b.RepeatMultiplier != c.wantRepe {
				t.Errorf("repeat mult = %.2f, want %.2f", b.RepeatMultiplier, c.wantRepe)
			}
			if b.FinalAmount != c.wantFinal {
				t.Errorf("final = %d, want %d", b.FinalAmount, c.wantFinal)
			}
		})
	}
}

func TestCalculate_UnknownTypeErrors(t *testing.T) {
	r := DefaultRuleset()
	if _, err := r.Calculate(Input{ViolationType: "jaywalking", OccurredAt: at(10)}); err == nil {
		t.Fatal("expected an error for an unknown violation type")
	}
}

func TestCalculate_RoundsHalfUp(t *testing.T) {
	// A custom ruleset that produces a fractional rupiah before rounding:
	// 333 * 1.0 * 1.5 = 499.5 -> rounds to 500.
	r := Ruleset{
		BaseAmounts:    map[string]int64{"x": 333},
		TimeMultiplier: TimeMultiplier{DayStartHour: 6, NightStartHour: 22, DayMultiplier: 1.0, NightMultiplier: 1.0},
		RepeatMultiplier: RepeatMultiplier{Tiers: []RepeatTier{
			{MinPriorUnpaid: 0, Multiplier: 1.0},
			{MinPriorUnpaid: 1, Multiplier: 1.5},
		}},
	}
	b, err := r.Calculate(Input{ViolationType: "x", OccurredAt: at(10), PriorUnpaidCount: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.FinalAmount != 500 {
		t.Errorf("final = %d, want 500 (499.5 rounded half-up)", b.FinalAmount)
	}
}

func TestValidate(t *testing.T) {
	if err := DefaultRuleset().Validate(); err != nil {
		t.Fatalf("default ruleset should be valid: %v", err)
	}

	missingType := DefaultRuleset()
	delete(missingType.BaseAmounts, "disabled_spot")
	if err := missingType.Validate(); err == nil {
		t.Error("expected error when a violation type is missing")
	}

	negativeBase := DefaultRuleset()
	negativeBase.BaseAmounts["expired_meter"] = -1
	if err := negativeBase.Validate(); err == nil {
		t.Error("expected error for non-positive base amount")
	}

	noTiers := DefaultRuleset()
	noTiers.RepeatMultiplier.Tiers = nil
	if err := noTiers.Validate(); err == nil {
		t.Error("expected error when there are no repeat tiers")
	}
}

func TestCalculate_SnapshotIsSelfContained(t *testing.T) {
	// The breakdown must carry every input to the formula so history can show
	// the fine exactly as issued, independent of later rule changes.
	r := DefaultRuleset()
	b, err := r.Calculate(Input{ViolationType: "disabled_spot", OccurredAt: at(23), PriorUnpaidCount: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.BaseAmount == 0 || b.TimeMultiplier == 0 || b.RepeatMultiplier == 0 || b.FinalAmount == 0 {
		t.Errorf("breakdown missing a component: %+v", b)
	}
	if b.PriorUnpaidCount != 2 {
		t.Errorf("prior unpaid count not captured: %+v", b)
	}
}
