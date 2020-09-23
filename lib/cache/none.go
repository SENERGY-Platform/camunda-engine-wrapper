package cache

import "encoding/json"

var None = &NoneCache{}

type NoneCache struct{}

func (this *NoneCache) Use(key string, getter func() (interface{}, error), result interface{}) (err error) {
	temp, err := getter()
	if err != nil {
		return err
	}
	value, err := json.Marshal(temp)
	if err != nil {
		return err
	}
	return json.Unmarshal(value, &result)
}

func (this *NoneCache) Invalidate(key string) (err error) {
	return nil
}
