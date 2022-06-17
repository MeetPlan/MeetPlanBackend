/// This file is a part of MeetPlan Proton, which is a part of MeetPlanBackend (https://github.com/MeetPlan/MeetPlanBackend).
///
/// Copyright (c) 2022, Mitja Ševerkar <mytja@protonmail.com> and The MeetPlan Team.
/// All rights reserved.
/// Use of this source code is governed by the GNU GPLv3 license, that can be found in the LICENSE file.

package proton

import (
	"encoding/json"
	"os"
)

const TeacherObjectType = "Teacher"
const HourObjectType = "Hour"
const DayObjectType = "Day"
const SubjectObjectType = "Subject"

const ProtonConfigVersion = "1.0"

type ProtonObject struct {
	ObjectID int    `json:"object_id"`
	Type     string `json:"type"`
}

type ProtonRule struct {
	Objects  []ProtonObject `json:"objects"`
	RuleName string         `json:"rule_name"`
	RuleType int            `json:"rule_type"`
}

type ProtonConfig struct {
	Version string       `json:"version"`
	Rules   []ProtonRule `json:"rules"`
}

func LoadConfig() (config ProtonConfig, err error) {
	file, err := os.ReadFile("protonConfig.json")
	if err != nil {
		marshal, err := json.Marshal(ProtonConfig{
			Version: ProtonConfigVersion,
			Rules:   make([]ProtonRule, 0),
		})
		if err != nil {
			return config, err
		}
		err = os.WriteFile("protonConfig.json", marshal, 0600)
		if err != nil {
			return config, err
		}
		file, err = os.ReadFile("protonConfig.json")
		if err != nil {
			return config, err
		}
	}
	err = json.Unmarshal(file, &config)
	return config, err
}

func SaveConfig(protonConfig ProtonConfig) error {
	marshal, err := json.Marshal(&protonConfig)
	if err != nil {
		return err
	}
	f, err := os.Create("protonConfig.json")
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	err = os.WriteFile("protonConfig.json", marshal, 0600)
	if err != nil {
		return err
	}
	return nil
}

func AddNewRule(protonConfig ProtonConfig, rule ProtonRule) (ProtonConfig, error) {
	protonConfig.Rules = append(protonConfig.Rules, rule)
	return protonConfig, SaveConfig(protonConfig)
}
