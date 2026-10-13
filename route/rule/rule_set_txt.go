package rule

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"

	"github.com/sagernet/fswatch"
	"github.com/sagernet/sing-box/adapter"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/logger"
	"github.com/sagernet/sing/common/x/list"
	"github.com/sagernet/sing/service/filemanager"

	"go4.org/netipx"
)

var _ adapter.RuleSet = (*TxtRuleSet)(nil)

type TxtRuleSet struct {
	ctx        context.Context
	logger     logger.Logger
	tag        string
	access     sync.RWMutex
	set        *Set
	metadata   adapter.RuleSetMetadata
	fileFormat string
	watcher    *fswatch.Watcher
	callbacks  list.List[adapter.RuleSetUpdateCallback]
	refs       atomic.Int32
}

func NewTxtRuleSet(ctx context.Context, logger logger.Logger, tag string, options option.RuleSet) (*TxtRuleSet, error) {
	ruleSet := &TxtRuleSet{
		ctx:        ctx,
		logger:     logger,
		tag:        tag,
		fileFormat: options.Format,
	}
	filePath := filemanager.BasePath(ctx, strings.ReplaceAll(options.LocalOptions.Path, C.RuleSetTagPlaceholder, tag))
	filePath, _ = filepath.Abs(filePath)
	err := ruleSet.reloadFile(filePath, logger)
	if err != nil {
		return nil, err
	}
	watcher, err := fswatch.NewWatcher(fswatch.Options{
		Path: []string{filePath},
		Callback: func(path string) {
			uErr := ruleSet.reloadFile(path, logger)
			if uErr != nil {
				logger.Error(E.Cause(uErr, "reload rule-set ", tag))
			}
		},
	})
	if err != nil {
		return nil, err
	}
	ruleSet.watcher = watcher
	return ruleSet, nil
}

func (s *TxtRuleSet) Name() string {
	return s.tag
}

func (s *TxtRuleSet) String() string {
	return "TxtRuleSet"
}

func (s *TxtRuleSet) StartContext(ctx context.Context, startContext *adapter.HTTPStartContext) error {
	if s.watcher != nil {
		err := s.watcher.Start()
		if err != nil {
			s.logger.Error(E.Cause(err, "watch rule-set file"))
		}
	}
	return nil
}

func (s *TxtRuleSet) reloadFile(path string, logger logger.Logger) error {
	f, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	var strs []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		line = strings.TrimFunc(line, func(r rune) bool {
			return !unicode.IsGraphic(r)
		})
		if line == "" {
			continue
		}
		strs = append(strs, line)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if len(strs) == 0 {
		return nil
	}
	set := NewSet(strs)
	logger.Warn(fmt.Sprintf("reloading %s, %d lines, set Size %.3fMB", path, len(strs), float64(set.Size())/(1024*1024)))
	return s.reloadRules(set)
}

func (s *TxtRuleSet) reloadRules(set *Set) error {
	var err error
	metadata := adapter.RuleSetMetadata{}
	err = validateRuleSetMetadataUpdate(s.ctx, s.tag, metadata)
	if err != nil {
		return err
	}
	s.access.Lock()
	s.set = set
	s.metadata = metadata
	callbacks := s.callbacks.Array()
	s.access.Unlock()
	for _, callback := range callbacks {
		callback(s)
	}
	return nil
}

func (s *TxtRuleSet) Metadata() adapter.RuleSetMetadata {
	s.access.RLock()
	defer s.access.RUnlock()
	return s.metadata
}

func (s *TxtRuleSet) ExtractIPSet() []*netipx.IPSet {
	s.access.RLock()
	defer s.access.RUnlock()
	return nil
}

func (s *TxtRuleSet) IncRef() {
	s.refs.Add(1)
}

func (s *TxtRuleSet) DecRef() {
	if s.refs.Add(-1) < 0 {
		panic("rule-set: negative refs")
	}
}

func (s *TxtRuleSet) Cleanup() {
	if s.refs.Load() == 0 {
		s.set = nil
	}
}

func (s *TxtRuleSet) RegisterCallback(callback adapter.RuleSetUpdateCallback) *list.Element[adapter.RuleSetUpdateCallback] {
	s.access.Lock()
	defer s.access.Unlock()
	return s.callbacks.PushBack(callback)
}

func (s *TxtRuleSet) UnregisterCallback(element *list.Element[adapter.RuleSetUpdateCallback]) {
	s.access.Lock()
	defer s.access.Unlock()
	s.callbacks.Remove(element)
}

func (s *TxtRuleSet) Close() error {
	s.set = nil
	return common.Close(common.PtrOrNil(s.watcher))
}

func (s *TxtRuleSet) Match(metadata *adapter.InboundContext) bool {
	if s.set == nil {
		return false
	}
	return s.set.Has(metadata.Domain)
}
