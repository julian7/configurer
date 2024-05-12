package configurer

import (
	"log/slog"
	"os"
	"testing"
)

type fakeConfigTwo struct {
	Twenty string
}

type fakeConfigThree struct {
	Thirty    string
	Thirtyone fakeConfigThirtyone
}

type fakeConfigThirtyone struct {
	Threehundredandten string
}

type fakeConfig struct {
	One   string
	Two   fakeConfigTwo
	Three []fakeConfigThree
}

type fakeLoader struct {
	oldConfig  *fakeConfig
	newConfig  *fakeConfig
	newAlready bool
}

func (l *fakeLoader) Filename() string {
	return "fake"
}

func (l *fakeLoader) Load() (Configuration, error) {
	if l.newAlready {
		return l.newConfig, nil
	}
	l.newAlready = true
	return l.oldConfig, nil
}

func TestReadConfig(t *testing.T) {
	gen := func() *fakeConfig {
		return &fakeConfig{
			One: "one",
			Two: fakeConfigTwo{
				Twenty: "twenty",
			},
			Three: []fakeConfigThree{
				{
					Thirty: "thirty",
					Thirtyone: fakeConfigThirtyone{
						Threehundredandten: "threehundredandten",
					},
				},
			},
		}
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	loader := &fakeLoader{
		oldConfig: gen(),
		newConfig: gen(),
	}
	loader.newConfig.Three[0].Thirty = "30"

	cctrl, err := New(loader, logger)
	if err != nil {
		t.Error("creating controller: %w", err)
	}

	truefalse := map[bool]string{
		true:  "true",
		false: "false",
	}

	tt := []struct {
		name    string
		changed string
		first   bool
		second  bool
	}{
		{name: "all", changed: "*", first: true, second: false},
		{name: "one", changed: "One", first: true, second: false},
		{name: "two", changed: "Two", first: true, second: false},
		{name: "two*", changed: "Two.*", first: true, second: false},
		{name: "twenty", changed: "Two.Twenty", first: true, second: false},
		{name: "three", changed: "Three", first: true, second: false},
		{name: "three*", changed: "Three.*", first: true, second: true},
		{name: "thirty", changed: "Three.0.Thirty", first: true, second: true},
		{name: "thirtyone", changed: "Three.0.Thirtyone", first: true, second: false},
		{name: "threehundredandten", changed: "Three.0.Thirtyone.Threehundredandten", first: true, second: false},
	}

	t.Run("first", func(t *testing.T) {
		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				got := cctrl.IsChanged(tc.changed)
				if tc.first != got {
					t.Errorf("IsChanged(%q) wanted %s, got %s", tc.changed, truefalse[tc.first], truefalse[got])
				}
			})
		}
	})

	cctrl.readConfig()

	t.Run("second", func(t *testing.T) {
		for _, tc := range tt {
			t.Run(tc.name, func(t *testing.T) {
				got := cctrl.IsChanged(tc.changed)
				if tc.second != got {
					t.Errorf("IsChanged(%q) wanted %s, got %s", tc.changed, truefalse[tc.second], truefalse[got])
				}
			})
		}
	})
}
