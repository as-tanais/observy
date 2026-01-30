PROFILE_DIR := profiles
BASE_PROFILE := $(PROFILE_DIR)/base.pprof
RESULT_PROFILE := $(PROFILE_DIR)/result.pprof
PPROF_URL := http://localhost:6060/debug/pprof/allocs?seconds=60

$(PROFILE_DIR):
	mkdir -p $(PROFILE_DIR)

.PHONY: base
base: $(PROFILE_DIR)
	curl "$(PPROF_URL)" -o $(BASE_PROFILE)
	@echo "base profile $(BASE_PROFILE)"

.PHONY: result
result: $(PROFILE_DIR)
	curl "$(PPROF_URL)" -o $(RESULT_PROFILE)
	@echo "result prof $(RESULT_PROFILE)"

.PHONY: compare
compare: $(BASE_PROFILE) $(RESULT_PROFILE)
	go tool pprof -top -diff_base=$(BASE_PROFILE) $(RESULT_PROFILE)
