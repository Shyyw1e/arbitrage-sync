package redisqueue

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Shyyw1e/arbitrage-sync/internal/core/usecase"
	"github.com/Shyyw1e/arbitrage-sync/internal/infrastructure/db"
	"github.com/Shyyw1e/arbitrage-sync/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
)

var (
	userStore db.UserStatesStore
	dispatcher = newDispatcher()
)

func InitRedisQueue(store db.UserStatesStore) {
	userStore = store
}

type cmdType int

const (
	cmdStart cmdType = iota
	cmdStop
	cmdUpdate
	cmdShutdown
)

type cmd struct {
	typ cmdType
	min float64
	max float64
	bot *tgbotapi.BotAPI
	reply chan error
}

type worker struct {
	chatID 		int64
	cmdCh  		chan cmd
	running 	atomic.Bool
	min 		atomic.Value
	max			atomic.Value
	bot 		atomic.Value		//tgbotapi.BotAPI
	hb 			atomic.Value		//time.Time (lastTick)
}

func (w *worker) getMin() float64            { v, _ := w.min.Load().(float64); return v }
func (w *worker) getMax() float64            { v, _ := w.max.Load().(float64); return v }
func (w *worker) getBot() *tgbotapi.BotAPI   { v, _ := w.bot.Load().(*tgbotapi.BotAPI); return v }
func (w *worker) lastHB() time.Time          { v, _ := w.hb.Load().(time.Time); return v }
func (w *worker) setHB(t time.Time)          { w.hb.Store(t) }
func (w *worker) isRunning() bool            { return w.running.Load() }
func (w *worker) setRunning(v bool)          { w.running.Store(v) }

func (w *worker) run(store db.UserStatesStore) {
	var (
		ticker *time.Ticker
		tickC <-chan time.Time
		//cancel context.CancelFunc
	)

	for {
		select {
		case c := <-w.cmdCh:
			switch c.typ {
			case cmdStart:
				w.min.Store(c.min)
				w.max.Store(c.max)
				w.bot.Store(c.bot)

				w.setHB(time.Now())

				if !w.isRunning() {
					logger.Log.Infof("starting worker %d", w.chatID)
					ticker = time.NewTicker(20*time.Second)
					tickC = ticker.C
					w.setRunning(true)
				}
				if c.reply != nil {
					c.reply <- nil
				}
			case cmdUpdate:
				logger.Log.Infof("updating worker %d", w.chatID)
				w.min.Store(c.min)
				w.max.Store(c.max)
				if c.bot != nil {
					w.bot.Store(c.bot)
				}

				if c.reply != nil {
					c.reply <- nil
				}

			case cmdStop:
				if w.isRunning() {
					w.setRunning(false)
					logger.Log.Infof("stopping worker %d", w.chatID)
					if ticker != nil {
						ticker.Stop()
					}
					tickC = nil
				}
				w.hb.Store(time.Time{})
				if c.reply != nil {
					c.reply <- nil
				}

			case cmdShutdown:
				if w.isRunning() {
					w.setRunning(false)
					if ticker != nil {
						ticker.Stop()
					}
					tickC = nil
					// if cancel != nil {
					// 	cancel()
					// }
				}
				if c.reply != nil {
					c.reply<-nil
				}
				return
			}
		case <-tickC:
			st, err := store.Get(w.chatID)
			if err != nil {
				logger.Log.WithError(err).Warnf("worker %d: failed to get userStore", w.chatID)
				continue
			}
			if st == nil || st.Step != "ready_to_run" {
				step := "<nil>"
				if st != nil { step = st.Step }
				logger.Log.Warnf("worker %d: inactive user state (step=%s)", w.chatID, step)
				continue
			}


			min, max := w.getMin(), w.getMax()
			bot := w.getBot()
			if bot == nil {
				logger.Log.Warnf("worker %d: bot is nil, skip tick", w.chatID)
				continue
			}

			facts, err := usecase.DetectFact(min, max, w.chatID)
			if err != nil {
				logger.Log.WithError(err).Warnf("worker %d: DetectFact failed, skip tick", w.chatID)
				continue
			}
			if len(facts) > 0 {
				for _, op := range facts {
					text := fmt.Sprintf("💰 Найден фактический арбитраж!\nBuy %s @ %.2f\nSell %s @ %.2f\nProfit: %.2f",
						op.BuyExchange, op.BuyPrice, op.SellExchange, op.SellPrice, op.ProfitMargin)
					_, err = bot.Send(tgbotapi.NewMessage(w.chatID, text))
					if err != nil {
						logger.Log.WithError(err).Warnf("worker %d: failed to send message", w.chatID)
					}
					time.Sleep(1500 * time.Millisecond)
				}
			}

			ops, pots, err := usecase.DetectAS(min, max, w.chatID)
			if err != nil {
				logger.Log.WithError(err).Warnf("worker %d: DetectAS failed", w.chatID)
				goto afterTick
			}
			if len(ops) > 0 {
				for _, op := range ops {
					text := fmt.Sprintf("💰 Найден потенциальный арбитраж!\nBuy %s @ %.2f\nSell %s @ %.2f\nProfit: %.2f",
						op.BuyExchange, op.BuyPrice, op.SellExchange, op.SellPrice, op.ProfitMargin)
					_, err = bot.Send(tgbotapi.NewMessage(w.chatID, text))
					if err != nil {
						logger.Log.WithError(err).Warnf("worker %d: failed to send message", w.chatID)
					}
					time.Sleep(1500 * time.Millisecond)
				}
			}
			if len(pots) > 0 {
				for _, op := range pots {
					text := fmt.Sprintf("💰 Найден обратный потенциальный арбитраж!\nBuy %s @ %.2f\nSell %s @ %.2f\nProfit: %.2f",
						op.BuyExchange, op.BuyPrice, op.SellExchange, op.SellPrice, op.ProfitMargin)
					_, err = bot.Send(tgbotapi.NewMessage(w.chatID, text))
					if err != nil {
						logger.Log.WithError(err).Warnf("worker %d: failed to send message", w.chatID)
					}
					time.Sleep(1500 * time.Millisecond)
				}
			}

			afterTick:
			// HB только после завершения тика (чтобы watchdog не трогал долгие парсы)
			w.setHB(time.Now())


		}
	}
}





type dispatcherT struct {
	mu 			sync.Mutex
	workers 	map[int64]*worker
}

func newDispatcher() *dispatcherT {
	return &dispatcherT{workers: make(map[int64]*worker)}
}

func (d *dispatcherT) ensure(chatID int64, store db.UserStatesStore) *worker {
	d.mu.Lock()
	defer d.mu.Unlock()

	if w, ok := d.workers[chatID]; ok {
		logger.Log.Infof("Worker %d: ensured", chatID)
		return w
	}

	w := &worker{chatID: chatID,
		cmdCh: make(chan cmd, 16),
	}

	w.hb.Store(time.Time{})
	d.workers[chatID] = w
	go w.run(store)

	logger.Log.Infof("Worker %d: ensured", chatID)
	return w
}

func (d *dispatcherT) start(chatID int64, min, max float64, bot *tgbotapi.BotAPI, store db.UserStatesStore) error {
	w := d.ensure(chatID, store)
	reply := make(chan error, 1)
	w.cmdCh<-cmd{typ: cmdStart, min: min, max: max, bot: bot, reply: reply}
	return <-reply
}

func (d *dispatcherT) stop(chatID int64, _ db.UserStatesStore) error {
	d.mu.Lock()
	w, ok := d.workers[chatID]
	d.mu.Unlock()
	if !ok {
		logger.Log.Warn("Nothing to stop, worker wasn't created")
		return nil
	}
	reply := make(chan error, 1)
	w.cmdCh<-cmd{typ: cmdStop, reply: reply}
	
	return <-reply
}

func (d *dispatcherT) isRunning(chatID int64) bool {
	d.mu.Lock()
	w, ok := d.workers[chatID]
	d.mu.Unlock()
	if !ok {
		return false
	}
	return w.isRunning()
}

func (d *dispatcherT) list() []*worker {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]*worker, 0, len(d.workers))
	for _, w := range d.workers {
		out = append(out, w)
	}
	return out
}

func StartWorkerLoop(bot *tgbotapi.BotAPI) {
	go func ()  {
		for {
			res, err := getRedis().BLPop(context.Background(), 10*time.Second, JobQueueKey).Result()
			if err != nil {
				if err == redis.Nil { // таймаут ожидания — норм
					continue
				}
				if isReadOnlyErr(err) {
					logger.Log.Warn("Redis returned READONLY on BLPOP — reconnecting...")
					_ = resetRedisClient()
					time.Sleep(2 * time.Second) // небольшой бэкофф
					continue
				}
				logger.Log.Errorf("BLPOP error: %v", err)
				time.Sleep(1 * time.Second)
				continue
			}

			if len(res) < 2 {
				continue
			}
			job := res[1]

			if !strings.HasPrefix(job, "detect-as:") {
				logger.Log.Warnf("unknown job: %s", job)
				continue
			}

			parts := strings.Split(job, ":")
			if len(parts) != 4 {
				logger.Log.Warnf("invalid job format: %s", job)
				continue
			}

			min, err1 := strconv.ParseFloat(parts[1], 64)
			max, err2 := strconv.ParseFloat(parts[2], 64)
			chatID, err3 := strconv.ParseInt(parts[3], 10, 64)
			if err1 != nil || err2 != nil || err3 != nil {
				logger.Log.Warnf("job parse error: %v %v %v (%s)", err1, err2, err3, job)
				continue
			}
			
			st, _ := userStore.Get(chatID)
			if st == nil || st.Step != "ready_to_run" {
				step := "<nil>"
				if st != nil { step = st.Step }
				logger.Log.Warnf("worker %d: inactive user state (step=%s)", chatID, step)
				continue
			}

			if err := dispatcher.start(chatID, min, max, bot, userStore); err != nil {
				logger.Log.Errorf("dispatcher start failed for %d: %v", chatID, err)
			}
		}
	}()

	go func() {
		t := time.NewTicker(15*time.Second)
		defer t.Stop()

		for range t.C {
			now := time.Now()
			ws := dispatcher.list()

			for _, w := range ws {
				if !w.isRunning() {
					continue
				}

				hb := w.lastHB()
				if hb.IsZero() || now.Sub(hb) <= 90*time.Second {
					continue
				}

				logger.Log.Warnf("Watchdog: stale worker chat=%d lastHB=%v -> soft restart", w.chatID, hb)

				if err := dispatcher.stop(w.chatID, userStore); err != nil {
					logger.Log.WithError(err).Warnf("Watchdog stop failed chat=%d", w.chatID)
					continue
				}

				if st, err := userStore.Get(w.chatID); err == nil && st != nil && st.Step == "ready_to_run" {
					if err := dispatcher.start(w.chatID, st.MinDiff, st.MaxSum, w.getBot(), userStore); err != nil {
						logger.Log.WithError(err).Warnf("Watchdog start failed chat=%d", w.chatID)
					}
				} else {
					logger.Log.Infof("Watchdog: chat=%d not active -> skip restart", w.chatID)
				}
			}
		}
	}()
}