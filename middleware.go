package pnet

import "github.com/luxingwen/pnet/utils"

func (nn *PNet) ApplyMiddleware(mw interface{}) error {
	applied := false
	errs := utils.NewErrors()

	err := nn.GetLocalNode().ApplyMiddleware(mw)
	if err == nil {
		applied = true
	} else {
		errs = append(errs, err)
	}

	for _, router := range nn.GetRouters() {
		err = router.ApplyMiddleware(mw)
		if err == nil {
			applied = true
		} else {
			errs = append(errs, err)
		}
	}

	if !applied {
		return errs.Merged()
	}

	return nil
}
