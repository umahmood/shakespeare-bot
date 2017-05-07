package shakespearebot

import (
	"math/rand"
	"sync"
	"time"
)

// random private random source
var random *rand.Rand

func init() {
	random = rand.New(
		&lockedRandSource{
			src: rand.New(rand.NewSource(time.Now().UnixNano())),
		},
	)
}

// lockedRandSource locked to prevent concurrent use of the underlying source
type lockedRandSource struct {
	lock sync.Mutex
	src  *rand.Rand
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64 from
// the default Source.
func (r *lockedRandSource) Int63() int64 {
	r.lock.Lock()
	ret := r.src.Int63()
	r.lock.Unlock()
	return ret
}

// Intn returns, as an int, a non-negative pseudo-random number in [0,n) from
// the default Source. It panics if n <= 0.
func (r *lockedRandSource) Intn(n int) int {
	r.lock.Lock()
	ret := r.src.Intn(n)
	r.lock.Unlock()
	return ret
}

// Seed uses the provided seed value to initialize the generator to a
// deterministic state.
func (r *lockedRandSource) Seed(seed int64) {
	r.lock.Lock()
	r.src.Seed(seed)
	r.lock.Unlock()
}
