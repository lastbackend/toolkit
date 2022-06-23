/*
Copyright [2014] - [2022] The Last.Backend authors.

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

package backoff

import (
	"math"
	"sync/atomic"
	"time"
)

const maxInt64 = float64(math.MaxInt64 - 512)

type Backoff struct {
	attempt uint64
	// Factor are the multiplication factor
	Factor float64
	// Min and Max are the minimum and maximum counter values.
	Min, Max time.Duration
}

func (b *Backoff) Duration() time.Duration {
	d := b.ForAttempt(float64(atomic.AddUint64(&b.attempt, 1) - 1))
	return d
}

func (b *Backoff) ForAttempt(attempt float64) time.Duration {
	min := b.Min
	if min <= 0 {
		min = 100 * time.Millisecond
	}
	max := b.Max
	if max <= 0 {
		max = 10 * time.Second
	}
	if min >= max {
		return max
	}

	factor := b.Factor
	if factor <= 0 {
		factor = 2
	}

	duration := float64(min) * math.Pow(factor, attempt)

	if duration > maxInt64 {
		return max
	}

	d := time.Duration(duration)
	if d < min {
		return min
	}
	if d > max {
		return max
	}
	return d
}

func (b *Backoff) Reset() {
	atomic.StoreUint64(&b.attempt, 0)
}

func (b *Backoff) Attempt() float64 {
	return float64(atomic.LoadUint64(&b.attempt))
}
