package parser

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

var (
	allocCtx    context.Context
	cancelAlloc context.CancelFunc

	startOnce sync.Once
	mu        sync.RWMutex

	sem       = make(chan struct{}, 1)
	defaultTO = 40 * time.Second       
)


func StartChromeAllocator() error {
	var startErr error
	startOnce.Do(func() {
		opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),       // или true, если твой Chrome старее
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-software-rasterizer", true),
	)
		allocCtx, cancelAlloc = chromedp.NewExecAllocator(context.Background(), opts...)
		ctx, cancel := chromedp.NewContext(allocCtx)
		defer cancel()
		_ = chromedp.Run(ctx)
	})
	return startErr
}

func StopChromeAllocator() {
	mu.Lock()
	defer mu.Unlock()
	if cancelAlloc != nil {
		cancelAlloc()
		cancelAlloc = nil
	}
}

func SetChromeParallelLimit(n int) {
	if n < 1 {
		n = 1
	}
	mu.Lock()
	defer mu.Unlock()
	sem = make(chan struct{}, n)
}

func NewTab() (context.Context, context.CancelFunc, error) {
	if allocCtx == nil {
		return nil, nil, errors.New("chrome allocator is not started")
	}

	sem <- struct{}{}

	tab, cancel := chromedp.NewContext(allocCtx)
	tab, cancelTO := context.WithTimeout(tab, defaultTO)

	cleanup := func() {
		cancelTO()
		cancel()
		select { 
		case <-sem:
		default:
		}
	}
	return tab, cleanup, nil
}

func RunSafe(ctx context.Context, actions ...chromedp.Action) error {
	EnsureAlive()
	backoff := []time.Duration{0, 500 * time.Millisecond, 2 * time.Second}
	var err error
	for i, d := range backoff {
		if d > 0 {
			time.Sleep(d)
		}
		err = chromedp.Run(ctx, actions...)
		if err == nil {
			return nil
		}
		if i == 0 {
			EnsureAlive()
		}
	}
	return err
}

func EnsureAlive() {
	mu.Lock()
	defer mu.Unlock()

	if allocCtx == nil {
		allocCtx, cancelAlloc = chromedp.NewExecAllocator(context.Background(), defaultOptions()...)
		warmup()
		return
	}

	pctx, pcancel := chromedp.NewContext(allocCtx)
	defer pcancel()
	tctx, tcancel := context.WithTimeout(pctx, 5*time.Second)
	defer tcancel()

	if err := chromedp.Run(tctx, chromedp.Navigate("about:blank")); err == nil {
		return
	}

	if cancelAlloc != nil {
		cancelAlloc()
	}
	allocCtx, cancelAlloc = chromedp.NewExecAllocator(context.Background(), defaultOptions()...)
	warmup()
}


func defaultOptions() []chromedp.ExecAllocatorOption {
	return append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", "new"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-zygote", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("mute-audio", true),
		chromedp.Flag("window-size", "1280,900"),
		chromedp.Flag("renderer-process-limit", 2),
		chromedp.Flag("js-flags", "--max-old-space-size=128"),
	)
}

func warmup() {
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	_ = chromedp.Run(ctx)
}

func EnsureNoOverlay(ctx context.Context) error {
	var hasTop bool
	_ = chromedp.Evaluate(`!!document.querySelector('dialog[open], [popover]:popover-open')`, &hasTop).Do(ctx)
	if hasTop {
		_ = chromedp.Click(`//button[contains(.,'Принять') or contains(.,'Accept') or contains(.,'OK')]`, chromedp.BySearch).Do(ctx)
		_ = chromedp.KeyEvent("\u001b").Do(ctx)
		_ = chromedp.Click(`//button[contains(@class,'close') or contains(@aria-label,'закрыть') or contains(@aria-label,'close') or .='×']`, chromedp.BySearch).Do(ctx)
		time.Sleep(150 * time.Millisecond)
	}
	return nil
}

func runOnceWithNewTab(opTimeout time.Duration, actions ...chromedp.Action) error {
	ctx, cancel, err := NewTab()
	if err != nil {
		return err
	}
	defer cancel()

	opCtx, opCancel := context.WithTimeout(ctx, opTimeout)
	defer opCancel()

	return RunSafe(opCtx, actions...)
}

func AcceptCookies(ctx context.Context) error {
	if err := chromedp.Run(ctx, chromedp.WaitReady("body", chromedp.ByQuery)); err != nil {
		return err
	}

	sel := `//button[contains(., 'Принять файлы cookie')] |
            //button[contains(., 'Принять')] |
            //button[contains(., 'Accept')]`

	return RunSafe(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			var exists bool
			_ = chromedp.Evaluate(`(function(){
               const xp = `+strconv.Quote(sel)+`;
               const x = document.evaluate(xp,document,null,XPathResult.FIRST_ORDERED_NODE_TYPE,null).singleNodeValue;
               if(!x) return false;
               const r = x.getBoundingClientRect();
               return r.width>0 && r.height>0;
            })()`, &exists).Do(ctx)
			if !exists {
				return nil
			}
			if err := chromedp.Click(sel, chromedp.BySearch).Do(ctx); err != nil {
				return err
			}
			time.Sleep(200 * time.Millisecond)
			return nil
		}),
		chromedp.Sleep(200*time.Millisecond),
	)
}
