-include .env

.PHONY: all build test test-coverage clean deploy-sepolia fund help install snapshot update format anvil remove

DEFAULT_ANVIL_KEY := 0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80

help:
    @echo "Usage:"
    @echo " make deploy [ARGS=...]"

NETWORK_ARGS := --rpc-url http://localhost:8545 --private-key $(DEFAULT_ANVIL_KEY) --broadcast

# if --network sepolia is used, then use sepolia stuff, otherwise anvil stuff
ifeq ($(findstring --network sepolia,$(ARGS)),--network sepolia)
	NETWORK_ARGS := --rpc-url $(SEPOLIA_RPC_URL) --private-key $(PRIVATE_KEY) --broadcast --verify --etherscan-api-key $(ETHERSCAN_API_KEY) --legacy -vvvvv
endif

build :; forge build

test-coverage :; forge coverage

test :; forge test

install:; forge install Cyfrin/foundry-devops@0.1.0 --no-commit && forge install foundry-rs/forge-std@v1.7.0 --no-commit && forge install transmissions11/solmate@v6 --no-commit

# deploy-sepolia :
# 	@forge script script/DeployRaffle.s.sol:DeployRaffle --rpc-url $(SEPOLIA_RPC_URL) --private-key $(PRIVATE_KEY) --broadcast --verify --etherscan-api-key $(ETHERSCAN_API_KEY) -vvvv

deploy :
	@forge script script/DeployRaffle.s.sol:DeployRaffle $(NETWORK_ARGS)

anvil :; anvil -m 'test test test test test test test test test test test junk' --steps-tracing --block-time 1

snapshot :; forge snapshot


# Clean the repo
clean  :; forge clean

# Remove modules
remove :; rm -rf .gitmodules && rm -rf .git/modules/* && rm -rf lib/forge-std && rm -rf lib/foundry-devops && rm -rf lib/solmate && touch .gitmodules && git add . && git commit -m "modules"

update :; forge update

all :; make clean && make remove && make install && make build

format :; forge fmt