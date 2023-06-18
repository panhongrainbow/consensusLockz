package lockz

import "time"

func (locker *Locker) Extend(key string) (err error) {
	ticker := time.NewTicker(locker.Opts.ExtendPeriod)
	for {
		select {
		case <-ticker.C:
			_, _, err = locker.client.Session().Renew(locker.sessionID, nil)
			if err != nil {
				return
			}
			err = locker.Incr(key)
			if err != nil {
				return
			}
		case <-locker.release:
			err = locker.DestroySession()
			return
		}
	}

	return
}

func (locker *Locker) Cancel() (err error) {
	locker.release <- doneAndReleaseLock{}
	return
}
