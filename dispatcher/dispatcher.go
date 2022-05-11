package dispatcher

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

const deferDuration = 1

// Dispatcher - структура "диспетчер" занимается выдачей очередного id аккаунта, балансирует нагрузку за счет ожидания очередного account id
// содержит два набора каналов и параметров, для очередей выпуска и перевыпуска сертификатов соответственно.
type Dispatcher struct {
	timeoutWaitID        int // время в секундах ожидания очередного id из каналов, до достижения таймута
	sliceOnlyNewAccounts []string
	newAccountID         chan string
	renewAccountID       chan string
	cntNewAcc            int
	cntRenewAcc          int
}

// Create - "конструктор" *Dispatcher
func Create(timeout int, onlyNew []string, renew []string) (*Dispatcher, error) {

	lenNew := len(onlyNew)
	lenRenew := len(renew)

	if lenNew == 0 {
		return nil, errors.New("not found any accountID for [new] orders")
	}

	if lenRenew == 0 {
		return nil, errors.New("not found any accountID for [renew] orders")
	}

	d := Dispatcher{
		timeoutWaitID:  timeout,
		newAccountID:   make(chan string, lenNew),
		renewAccountID: make(chan string, lenRenew),
		cntNewAcc:      lenNew,
		cntRenewAcc:    lenRenew}

	// нужно для функции isNew()
	copy(d.sliceOnlyNewAccounts, onlyNew)

	for _, id := range onlyNew {
		d.newAccountID <- id
	}

	for _, id := range renew {
		d.renewAccountID <- id
	}

	return &d, nil
}

// NextNewAccountID - возвращает следующий аккаунт id доступный для выпуска новых сертов, и флаг - успех/нет
func (d *Dispatcher) NextNewAccountID(ctx context.Context) (string, bool) {

	ctxWithTimeout, cancelFunction := context.WithTimeout(ctx, time.Duration(d.timeoutWaitID)*time.Second)
	defer cancelFunction()

	select {
	case <-ctxWithTimeout.Done():
		return "", false
	case id := <-d.newAccountID:
		return id, true
	}
}

// NextRenewAccountID - возвращает следующий аккаунт id доступный для перевыпуска сертов, и флаг - успех/нет
func (d *Dispatcher) NextRenewAccountID(ctx context.Context) (string, bool) {

	ctxWithTimeout, cancelFunction := context.WithTimeout(ctx, time.Duration(d.timeoutWaitID)*time.Second)
	defer cancelFunction()

	select {
	case <-ctxWithTimeout.Done():
		return "", false
	case id := <-d.renewAccountID:
		return id, true
	}
}

// FreeNewAccountID - возвращает использованный id в канал new accounts id
func (d *Dispatcher) FreeNewAccountID(ctx context.Context, id string) error {

	ctxWithTimeout, cancelFunction := context.WithTimeout(ctx, time.Duration(deferDuration)*time.Second)
	defer cancelFunction()

	select {
	case <-ctxWithTimeout.Done():
		return fmt.Errorf("timeout to free [new] acc id: %s ,full chan [newAccountID] probably", id)
	case d.newAccountID <- id:
		return nil
	}
}

// FreeRenewAccountID - возвращает использованный id в канал renew accounts id
func (d *Dispatcher) FreeRenewAccountID(ctx context.Context, id string) error {

	ctxWithTimeout, cancelFunction := context.WithTimeout(ctx, time.Duration(deferDuration)*time.Second)
	defer cancelFunction()

	select {
	case <-ctxWithTimeout.Done():
		return fmt.Errorf("timeout to free [renew] acc id: %s ,full chan [renewAccountID] probably", id)
	case d.renewAccountID <- id:
		return nil
	}
}
