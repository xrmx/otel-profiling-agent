/*
 * Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
 * or more contributor license agreements. Licensed under the Apache License 2.0.
 * See the file "LICENSE" for details.
 */

package xsync_test

import (
	"errors"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/elastic/otel-profiling-agent/libpf/xsync"
)

func TestOnceLock(t *testing.T) {
	attempt := 0 // intentionally not atomic
	once := xsync.Once[string]{}
	someError := errors.New("oh no")
	numOk := atomic.Uint32{}
	wg := sync.WaitGroup{}

	assert.Nil(t, once.Get())

	for i := 0; i < 32; i++ {
		wg.Add(1)

		go func() {
			val, err := once.GetOrInit(func() (string, error) {
				if attempt == 3 {
					time.Sleep(25 * time.Millisecond)
					return strconv.Itoa(attempt), nil
				}

				attempt++
				return "", someError
			})

			switch err {
			case someError:
				assert.Nil(t, val)
			case nil:
				numOk.Add(1)
				assert.Equal(t, "3", *val)
			default:
				assert.Fail(t, "unreachable")
			}

			wg.Done()
		}()
	}

	wg.Wait()
	assert.Equal(t, "3", *once.Get())
	assert.Equal(t, uint32(32-3), numOk.Load())
}
