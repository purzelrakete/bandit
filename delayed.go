package bandit

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

// NewDelayedBandit wraps the given bandit.
func NewDelayedBandit(b Bandit, updates chan Counters) (Bandit, error) {
	bandit := delayedBandit{
		bandit:  b,
		updates: updates,
	}

	go func() {
		for counters := range updates {
			b.Init(&counters)
		}
	}()

	return &bandit, nil
}

// delayedBandit wraps a bandit. Internal counters are stored at the
// configured source file, which is pooled at `poll` interval. The retrieved
// Snapshot replaces the bandit's internal counters.
type delayedBandit struct {
	Counters
	updates chan Counters
	bandit  Bandit
}

// SelectArm delegates to the wrapped bandit
func (b *delayedBandit) SelectArm() int {
	return b.bandit.SelectArm()
}

// Version gives information about delayed bandit + the wrapped bandit.
func (b *delayedBandit) Version() string {
	return fmt.Sprintf("Delayed(%s)", b.bandit.Version())
}

// DelayedUpdate updates the internal counters of a bandit with the provided
// counters.
func (b *delayedBandit) Init(c *Counters) error {
	b.Lock()
	defer b.Unlock()
	return b.bandit.Init(c)
}

// Update is a NOP. Delayed bandit is updated with Reset(counter) instead
func (b *delayedBandit) Update(arm int, reward float64) {}

// Opener can be used to reopen underlying file descriptors.
type Opener interface {
	Open() (io.ReadCloser, error)
}

// NewHTTPOpener returns an opener using an underlying URL.
func NewHTTPOpener(url string) Opener {
	return &httpOpener{
		URL: url,
	}
}

type httpOpener struct {
	URL string
}

func (o *httpOpener) Open() (io.ReadCloser, error) {
	resp, err := http.Get(o.URL)
	if err != nil {
		return nil, fmt.Errorf("http GET failed: %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("http GET not 200: %s", resp.StatusCode)
	}
	return resp.Body, nil
}

// NewFileOpener returns an Opener using and underlying file.
func NewFileOpener(filename string) Opener {
	return &fileOpener{
		Filename: filename,
	}
}

type fileOpener struct {
	Filename string
}

func (o *fileOpener) Open() (io.ReadCloser, error) {
	reader, err := os.Open(o.Filename)
	if err != nil {
		return nil, err
	}

	return reader, err
}

// NewSimulatedDelayedBandit simulates delayed bandit by flushing counters to
// the underlying bandit after `flush` number of updates.
func NewSimulatedDelayedBandit(b Bandit, arms, flush int) Bandit {
	return &simulatedDelayedBandit{
		limit:   flush,
		updates: flush,
		delayedBandit: delayedBandit{
			bandit:   b,
			Counters: NewCounters(arms),
		},
	}
}

// simulatedDelayedBandit is used for testing but also simulation and
// plotting. It simulates delayed bandit by flushing counters to the
// underlying bandit after `limit` number of updates.
type simulatedDelayedBandit struct {
	delayedBandit
	limit   int // #updates to wait before flushing Counters to underlying bandit
	updates int // #updates since last flush
}

// Update flushes counters to the underlying bandit every n updates. This is
// approximately the behaviour seen by a delayed bandit in production.
func (b *simulatedDelayedBandit) Update(arm int, reward float64) {
	b.Lock()
	defer b.Unlock()

	arm--
	b.counts[arm]++
	count := b.counts[arm]
	b.values[arm] = ((b.values[arm] * float64(count-1)) + reward) / float64(count)

	b.updates++
	if b.updates >= b.limit {
		b.bandit.Init(&b.Counters)
		b.updates = 0
	}
}
