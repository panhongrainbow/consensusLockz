package lockz

import "time"

func (locker *Locker) Extend(key string) (err error) {
	ticker := time.NewTicker(locker.Opts.RenewPeriod)
	for {
		select {
		case <-ticker.C:
			_, _, err = locker.Client.Session().Renew(locker.SessionID, nil)
			if err != nil {
				return
			}
			err = locker.Incr(key)
			if err != nil {
				return
			}
		case <-locker.cancel:
			err = locker.DestroySession()
			return
		}
	}

	return
}

func (locker *Locker) Cancel() (err error) {
	locker.cancel <- done{}
	return
}
