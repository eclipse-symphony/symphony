/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

package memory

import (
	"encoding/json"
)

type Mode int

const (
	Dedicated Mode = iota
	Global
	Shared
	Unknown
)

func (s Mode) String() string {
	return [...]string{"Dedicated", "Global", "Shared"}[s]
}

func (s Mode) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *Mode) UnmarshalJSON(data []byte) error {
	var statusStr string
	if err := json.Unmarshal(data, &statusStr); err != nil {
		return err
	}

	switch statusStr {
	case "Dedicated":
		*s = Dedicated
	case "Global":
		*s = Global
	case "Shared":
		*s = Shared
	default:
		*s = Unknown
	}
	return nil
}

type MemoryKeyLockProviderConfig struct {
	CleanInterval int  `json:"cleanInterval"`
	PurgeDuration int  `json:"purgeDuration"`
	Mode          Mode `json:"mode"`
}
