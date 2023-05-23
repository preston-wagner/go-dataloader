package dataloader

func ErrForAll[KEY_TYPE comparable](keys []KEY_TYPE, err error) map[KEY_TYPE]error {
	errs := map[KEY_TYPE]error{}
	for _, key := range keys {
		errs[key] = err
	}
	return errs
}
