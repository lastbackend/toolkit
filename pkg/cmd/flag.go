/*
Copyright [2014] - [2023] The Last.Backend authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"time"
)

type Flags []Flag

func (f *Flags) AddStringFlag(name string, dest *string) *StringFlag {
	flag := &StringFlag{
		name:        name,
		destination: dest,
	}
	*f = append(*f, flag)
	return flag
}

func (f *Flags) AddIntFlag(name string, dest *int) *IntFlag {
	flag := &IntFlag{
		name:        name,
		destination: dest,
	}
	*f = append(*f, flag)
	return flag
}

func (f *Flags) AddInt32Flag(name string, dest *int32) *Int32Flag {
	flag := &Int32Flag{
		name:        name,
		destination: dest,
	}
	*f = append(*f, flag)
	return flag
}

func (f *Flags) AddBoolFlag(name string, dest *bool) *BoolFlag {
	flag := &BoolFlag{
		name:        name,
		destination: dest,
	}
	*f = append(*f, flag)
	return flag
}

func (f *Flags) AddStringSliceFlag(name string, dest *[]string) *StringSliceFlag {
	flag := &StringSliceFlag{
		name:        name,
		destination: dest,
	}
	*f = append(*f, flag)
	return flag
}

func (f *Flags) AddDurationFlag(name string, dest *time.Duration) *DurationFlag {
	flag := &DurationFlag{
		name:        name,
		destination: dest,
	}
	*f = append(*f, flag)
	return flag
}
