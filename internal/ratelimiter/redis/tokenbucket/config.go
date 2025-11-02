package tokenbucket

import "fmt"

type TBConfig struct {
	Capacity   int64 `json:"capacity"`
	RefillRate int64 `json:"refillRate"`
}

func (c *TBConfig) Validate() error {
	if c.Capacity <= 0 {
		return fmt.Errorf("token bucket capacity must be positive")
	}
	if c.RefillRate <= 0 {
		return fmt.Errorf("token bucket refill rate must be positive")
	}
	return nil
}
