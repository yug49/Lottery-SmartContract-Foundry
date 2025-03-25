//SPDX-License-Identifier: MIT

/**
 * @title A sample Raffle Contract
 * @author Yug Agarwal
 * @notice This contract is for creating a sample Raffle Contract
 * @dev  Implements Chainlink VRFv2.5 and Chainlink Automation
 */

// Layout of the contract file:
// version
// imports
// errors
// interfaces, libraries, contract

// Inside Contract:
// Type declarations
// State variables
// Events
// Modifiers
// Functions

// Layout of Functions:
// constructor
// receive function (if exists)
// fallback function (if exists)
// external
// public
// internal
// private

// view & pure functions

pragma solidity ^0.8.19;

import {VRFConsumerBaseV2Plus} from
    "../lib/chainlink/contracts/src/v0.8/vrf/dev/VRFConsumerBaseV2Plus.sol";
import {VRFV2PlusClient} from
    "../lib/chainlink/contracts/src/v0.8//vrf/dev/libraries/VRFV2PlusClient.sol";
import {AutomationCompatibleInterface} from
    "../lib/chainlink/contracts/src/v0.8/automation/interfaces/AutomationCompatibleInterface.sol";
import {IVRFCoordinatorV2Plus} from
    "../lib/chainlink/contracts/src/v0.8/vrf/dev/interfaces/IVRFCoordinatorV2Plus.sol";
import {VRFV2PlusClient} from
    "../lib/chainlink/contracts/src/v0.8/vrf/dev/libraries/VRFV2PlusClient.sol";

contract Raffle is VRFConsumerBaseV2Plus, AutomationCompatibleInterface {
    /**
     * Errors
     */
    error Raffle__SendMoreToEnterRaffle();
    error Raffle__TransferFailed();
    error Raffle__RaffleNotOpen();
    error Raffle__UpKeepNotNeeded(uint256 balance, uint256 playersLength, RaffleState raffleState);
    error Raffle__LotteryIsCalculatingWinner();

    /**
     * Type Declarations
     */
    enum RaffleState {
        OPEN, // index 0
        CALCULATING //index 1

    }

    /* State Variables */
    RaffleState private s_raffleState;
    IVRFCoordinatorV2Plus private immutable i_vrfCoordinatorContract;

    uint32 private constant NUM_WORDS = 1; //number of random numbers
    uint16 private constant REQUEST_CONFIRMATIONS = 3;

    uint256 private immutable i_enteranceFee;
    uint256 private immutable i_interval;
    /**
     * @dev duration of a lottery in seconds
     */
    bytes32 private immutable i_gasLane; //gas willing to pay
    uint256 private immutable i_subscriptionId;
    uint256 private s_lastTimeStamp;
    address payable[] private s_players;
    uint32 private immutable i_callBackGasLimit; //gas limit willing to generate
    address private s_recentWinner;

    /**
     * Events
     */
    event RaffleEntered(address indexed player);
    event WinnerPicked(address indexed winner);
    event RequestedRaffleWinner(uint256 indexed requestId);

    constructor(
        uint256 enteranceFee,
        uint256 interval,
        address vrfCoordinator,
        bytes32 gasLane,
        uint256 subsriptionId,
        uint32 callBackGasLimit
    ) VRFConsumerBaseV2Plus(vrfCoordinator) {
        //if parent has a constructor, always call it when calling the child's
        i_enteranceFee = enteranceFee;
        i_interval = interval;
        i_gasLane = gasLane;
        i_subscriptionId = subsriptionId;
        i_callBackGasLimit = callBackGasLimit;
        s_lastTimeStamp = block.timestamp; //current block timestamp
        s_raffleState = RaffleState.OPEN;
        i_vrfCoordinatorContract = IVRFCoordinatorV2Plus(vrfCoordinator);
    }

    function enterRaffle() external payable {
        //require(msg.value >= i_enteranceFee, "Not enough ETH to enter the raffle");   //not gas efficient
        if (msg.value < i_enteranceFee) {
            revert Raffle__SendMoreToEnterRaffle(); //in 0.8.26 and above we can revert in require, although using if-revert is most gas efficient
        }
        if (s_raffleState == RaffleState.CALCULATING) {
            revert Raffle__RaffleNotOpen();
        }
        s_players.push(payable(msg.sender));

        // Events:
        // 1. Makes migration easier
        // 2. Makes front-end "indexing" easier
        emit RaffleEntered(msg.sender); //every time you update storage, you will want to emit an eventw
    }

    /**
     * @dev This is the function that the Chainlink nodes will call to see if the lottery is ready to have a winner picked.
     * The following should be true in order for upKeepNeed to be true:
     * 1. The time interval has passed between raffle runs
     * 2. the lottery is open
     * 3. The contract has ETH
     * 4. There are players registered
     * 5. Implicitly, your subscription has ETH
     * @param - ignored
     * @return upKeepNeeded - true if it's time to restart the lottery
     * @return - ignoreed
     */
    function checkUpkeep(bytes memory /* checkData */ )
        public
        view
        override
        returns (bool upKeepNeeded, bytes memory /*performData*/ )
    {
        bool timeHasPassed = ((block.timestamp - s_lastTimeStamp) >= i_interval);
        bool isOpen = (s_raffleState == RaffleState.OPEN);
        bool hasBalance = address(this).balance > 0;
        bool hasPlayers = s_players.length > 0;
        upKeepNeeded = (timeHasPassed && isOpen && hasBalance && hasPlayers);

        return (upKeepNeeded, "0x0");
    }

    /**
     *
     * @dev Once `checkUpkeep` is returning `true`, this function is called and it kicks off a chainlinkVrf Call to get a random winner.
     */
    function performUpkeep(bytes calldata /* performData */ ) external override {
        // 1. Get a random number
        // 2. Use Random number to pick a player
        // 3. Be automatically called
        (bool upKeepNeeded,) = checkUpkeep("");
        if (!upKeepNeeded) {
            revert Raffle__UpKeepNotNeeded(address(this).balance, s_players.length, s_raffleState);
        }
        s_raffleState = RaffleState.CALCULATING;

        // Get our random number
        // 1. Request RNG
        // 2. Get RNG

        // uint256 requestId = s_vrfCoordinator.requestRandomWords(
        //     VRFV2PlusClient.RandomWordsRequest({
        //         keyHash: i_gasLane,
        //         subId: i_subscriptionId,
        //         requestConfirmations: REQUEST_CONFIRMATIONS,
        //         callbackGasLimit: i_callBackGasLimit,
        //         numWords: NUM_WORDS,
        //         extraArgs: VRFV2PlusClient._argsToBytes(
        //             // Set nativePayment to true to pay for VRF requests with Sepolia ETH instead of LINK
        //             VRFV2PlusClient.ExtraArgsV1({nativePayment: false})
        //         )
        //     })
        // );

        uint256 requestId = i_vrfCoordinatorContract.requestRandomWords(
            VRFV2PlusClient.RandomWordsRequest({
                keyHash: i_gasLane,
                subId: i_subscriptionId,
                requestConfirmations: REQUEST_CONFIRMATIONS,
                callbackGasLimit: i_callBackGasLimit,
                numWords: NUM_WORDS,
                // set nativePayment to true to pay the VRF requests with Sepolia ETH instead of LINK
                extraArgs: VRFV2PlusClient._argsToBytes(VRFV2PlusClient.ExtraArgsV1({nativePayment: false}))
            })
        );

        //Is this redundant? -> yes coz vrf also emits an event for this
        emit RequestedRaffleWinner(requestId);
    }

    // CEI: Checks, Effects, Interactions Pattern
    function fulfillRandomWords(uint256, /*requestId*/ uint256[] calldata randomWords) internal override {
        // Checks --> require, conditionals etc.

        // Effects (Internal Contract State)
        uint256 indexOfWinner = randomWords[0] % s_players.length;
        address payable recentWinner = s_players[indexOfWinner];
        s_recentWinner = recentWinner;

        s_players = new address payable[](0);
        s_lastTimeStamp = block.timestamp;
        emit WinnerPicked(s_recentWinner);

        s_raffleState = RaffleState.OPEN;

        // Interactions (External Contract Interactions)
        (bool success,) = recentWinner.call{value: address(this).balance}("");
        if (!success) {
            revert Raffle__TransferFailed();
        }
    }

    /* Getter Functions */

    function getEnteranceFee() external view returns (uint256) {
        return i_enteranceFee;
    }

    function getRaffleState() external view returns (RaffleState) {
        return s_raffleState;
    }

    function getPlayer(uint256 indexOfPlayer) external view returns (address) {
        return s_players[indexOfPlayer];
    }

    function getLastTimeStamp() external view returns (uint256) {
        return s_lastTimeStamp;
    }

    function getRecentWinner() external view returns (address) {
        return s_recentWinner;
    }

    function getNumWords() public pure returns (uint256) {
        return NUM_WORDS;
    }

    function getRequestConfirmation() public pure returns (uint256) {
        return REQUEST_CONFIRMATIONS;
    }

    function getInterval() public view returns (uint256) {
        return i_interval;
    }

    function getNumberOfPlayers() public view returns (uint256) {
        return s_players.length;
    }
}
