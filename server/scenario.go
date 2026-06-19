package server

import "github.com/Saku0512/specter/config"

func applyScenarioPreset(name string, preset config.Scenario, state *StateStore, vars *VarStore, scenario *ScenarioStore, store *DataStore) {
	state.Set(preset.State)
	vars.Replace(preset.Vars)
	store.ReplaceAll(preset.Stores)
	scenario.Set(name)
}
