package main

import (
	"bufio"
	"fmt"
	"obcsdk/chaincode"
	"obcsdk/peernetwork"
	"os"
	"strconv"
	"strings"
	"time"
	"math/rand"
)

var f *os.File
var writer *bufio.Writer
var myNetwork peernetwork.PeerNetwork
var url string
var overallTestPass bool

func main() {

	overallTestPass = true
	var err error
	f, err = os.OpenFile("/tmp/hyperledgerBetaTestrun_Output", os.O_RDWR|os.O_APPEND, 0660)
        if ( err != nil) {
          fmt.Println("Output file does not exist creating one ..")
          f, err = os.Create("/tmp/hyperledgerBetaTestrun_Output")
        }
	//check(err)
	defer f.Close()
	writer = bufio.NewWriter(f)

	myStr := fmt.Sprintf("\n\n*********** BEGIN chcotest/BasicFuncExistingNetwork.go ***************")
	fmt.Println(myStr)
	fmt.Fprintln(writer, myStr)

	defer timeTrack(time.Now(), "Testcase execution Done")

	setupNetwork()

	fmt.Println("\n===== Start userRegisterTest =====")
	//get a URL details to get info n chainstats/transactions/blocks etc.
	//aPeer, _ := peernetwork.APeer(myNetwork)
	//url = "http://" + aPeer.PeerDetails["ip"] + ":" + aPeer.PeerDetails["port"]

	user_ip, user_port, user_name, err := peernetwork.PeerOfThisUser(myNetwork, "test_user0")
	url = "http://" + user_ip + ":" + user_port
	userRegisterTest(url, user_name)

	fmt.Println("\n===== Start NetworkPeers Test =====")
	response, status := chaincode.NetworkPeers(url)
	if strings.Contains(status, "200") {
		myStr = fmt.Sprintf("NetworkPeers Rest API TEST PASS: successful.")
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
	} else {
		overallTestPass = false
		myStr = fmt.Sprintf("NetworkPeers Rest API TEST FAIL!!! response:\n%s\n", response)
		fmt.Println(response)
		fmt.Fprintln(writer, response)
	}

	fmt.Println("\n===== ChainStats Test =====")
	response, status = chaincode.GetChainStats(url)
	height := chaincode.Monitor_ChainHeight(url) // save the height; it will be needed below for getHeight()
	if strings.Contains(status, "200") {
		myStr = fmt.Sprintf("ChainStats Rest API TEST PASS.")
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
	} else {
		overallTestPass = false
		myStr = fmt.Sprintf("ChainStats Rest API TEST FAIL!!!")
		fmt.Println(response)
		fmt.Fprintln(writer, response)
	}


	fmt.Println("\n===== Deploy Test =====")

	// To ensure we deploy new values, we should randomize the values used.
	// Remember: the deploy will get ignored if we previously deployed using same values,
	// and this deploy test would fail if we run this testcase multiple times on same network.

	var inita, initb, curra, currb int

	// generate random values for "a" and "b" between 0-9999
	// use NewSource instead of simply rand.Intn(10000), so it is different each time the program runs

	s1 := rand.NewSource(time.Now().UnixNano()) // generate a new seed for "a"
	r1 := rand.New(s1)
	inita = r1.Intn(10000)

	s2 := rand.NewSource(time.Now().UnixNano()) // generate a new seed for b
	r2 := rand.New(s2)
	initb = r2.Intn(10000)

	curra = inita
	currb = initb

	deploy(inita,initb)		// deploy sleeps for 30 secs too
	//time.Sleep(30000 * time.Millisecond)

	query("DEPLOY", curra, currb)

	fmt.Println("\n===== Invoke Test =====")
	invRes := invoke()
	time.Sleep(10000 * time.Millisecond)
	curra = curra - 1
	currb = currb + 1

	query("INVOKE", curra, currb)

	fmt.Println("\n===== GetChainHeight Test =====")
	height = height + 2    // the original height, plus 2 for the deploy and invoke transactions
	getHeight(height) // verify all peers match this expected height

	fmt.Println("\n===== GetBlockStats API Test =====")
	//chaincode.BlockStats(url, height)
	nonHashData, _ := chaincode.GetBlockTrxInfoByHost("PEER0", height-1)

	if strings.Contains(nonHashData.TransactionResult[0].Uuid, invRes) {
		myStr = fmt.Sprintf("\nGetBlocks API TEST PASS: Transaction Successfully stored in Block")
	} else {
		overallTestPass = false
		myStr = fmt.Sprintf("\nGetBlocks API TEST FAIL: Transaction NOT stored in Block")
	}
	myStr += fmt.Sprintf("\nCH_Block = %d, UUID = %s, InvokeTransactionResult = %s\n", height-1, nonHashData.TransactionResult[0].Uuid, invRes)
	fmt.Printf(myStr)
	fmt.Fprintf(writer, myStr)
	writer.Flush()

	fmt.Println("\n===== Get Transaction_Detail Test =====")
	fmt.Println("url:  " + url)
	fmt.Println("invRes:  " + invRes)
	fmt.Println("Transaction_Detail(url,invRes):  ")
	chaincode.Transaction_Detail(url, invRes)

	testResult := "PASS"
	if !overallTestPass { testResult = "FAIL" }
	myStr = fmt.Sprintf("\n\n*********** END chcotest/BasicFuncExistingNetwork.go OVERALL TEST RESULT = %s ***************\n\n", testResult)
	fmt.Println(myStr)
	fmt.Fprintln(writer, myStr)
	writer.Flush()
}

func setupNetwork() {
	//fmt.Println("Working with a new network: Setting up a local network with 4 peers with security")
	//peernetwork.SetupLocalNetwork(4, true)

	fmt.Println("===== Working with an existing network: Retrieving network information and connecting to it =====")

        fmt.Println("----- Get existing Network Credentials ----- ")
        peernetwork.GetNC_Local()

        fmt.Println("----- Connect to existing network - InitNetwork -----")
	myNetwork = chaincode.InitNetwork()

	fmt.Println("----- InitChainCodes -----")
	chaincode.InitChainCodes()
	time.Sleep(50000 * time.Millisecond)

	fmt.Println("----- RegisterUsers -----")
	chaincode.RegisterUsers()
	//peernetwork.PrintNetworkDetails(myNetwork)
	peernetwork.PrintNetworkDetails()
	numPeers := peernetwork.GetNumberOfPeers(myNetwork)

	myStr := fmt.Sprintf("Network running with %d peers with pbft and security+privacy enabled\n", numPeers)
	fmt.Println(myStr)
	fmt.Fprintln(writer, myStr)
}

// arg = a username that was already registered; this func confirms if it was successful
// and confirms user "ghostuserdoesnotexist" is not registered
// and confirms 
func userRegisterTest(url string, username string) {

	fmt.Println("\n----- RegisterUser Test -----")
	response, status := chaincode.UserRegister_Status(url, username)
	myStr := "RegisterUser API TEST "
	if strings.Contains(status, "200") && strings.Contains(response, username + " is already logged in") {
		myStr += fmt.Sprintf ("PASS: %s User Registration was already done successfully", username)
	} else {
		overallTestPass = false
		myStr += fmt.Sprintf ("FAIL: %s User Registration was NOT already done\n status = %s\n response = %s", username, status, response)
	}
	fmt.Println(myStr)
	time.Sleep(1000 * time.Millisecond)

	fmt.Println("\n----- RegisterUser Negative Test -----")
	response, status = chaincode.UserRegister_Status(url, "ghostuserdoesnotexist")
	if ((strings.Contains(status, "200")) == false) {
		fmt.Println("RegisterUser API Negative TEST PASS: CONFIRMED that user <ghostuserdoesnotexist> is unregistered as expected")
	} else {
		overallTestPass = false
		myStr = fmt.Sprintf("RegisterUser API Negative TEST FAIL: User <ghostuserdoesnotexist> was found in Registrar User List but it was never registered!\n status = %s\n response = %s\n", status, response)
		fmt.Println(myStr)
	}
	time.Sleep(1000 * time.Millisecond)

 /*
	fmt.Println("\n----- UserRegister_ecert Test -----")
	ecertUser := "lukas"
	response, status = chaincode.UserRegister_ecertDetail(url, ecertUser)
	myStr = "\nUserRegister_ecert TEST "
	if strings.Contains(status, "200") && strings.Contains(response, ecertUser + " is already logged in") {
		myStr += fmt.Sprintf ("PASS: %s ecert User Registration was already done successfully", ecertUser)
	} else {
		overallTestPass = false
		myStr += fmt.Sprintf ("FAIL: %s ecert User Registration was NOT already done\n status = %s\n response = %s\n", username, status, response)
	}
	fmt.Println(myStr)
	time.Sleep(1000 * time.Millisecond)
 */

	fmt.Println("\n----- UserRegister_ecert Negative Test -----")
	response, status = chaincode.UserRegister_ecertDetail(url, "ghostuserdoesnotexist")
	if ((strings.Contains(status, "200")) == false) {
		fmt.Println("UserRegister_ecert API Negative TEST PASS: CONFIRMED that user <ghostuserdoesnotexist> is unregistered as expected")
	} else {
		overallTestPass = false
		myStr = fmt.Sprintf("UserRegister_ecert API Negative TEST FAIL: User <ghostuserdoesnotexist> was found in Registrar User List but it was never registered!\n status = %s\n response = %s\n", status, response)
		fmt.Println(myStr)
	}

	time.Sleep(5000 * time.Millisecond)
}

func deploy(a_value, b_value int) {					// using example02
	fmt.Println("\nPOST/Chaincode: Deploying chaincode with values for a, b : ", a_value, b_value)
	dAPIArgs0 := []string{"example02", "init"}
	depArgs0 := []string{"a", strconv.Itoa(a_value), "b", strconv.Itoa(b_value)}
	chaincode.Deploy(dAPIArgs0, depArgs0)
	time.Sleep(30000 * time.Millisecond) // minimum delay required is 30, in local environment
}

func invoke() string {						// using example02
	fmt.Println("\nPOST/Chaincode: Invoke chaincode ....")
	iAPIArgs0 := []string{"example02", "invoke"}
	invArgs0 := []string{"a", "b", "1"}
	invRes, _ := chaincode.Invoke(iAPIArgs0, invArgs0)
	return invRes
}

func query(txName string, expectedA int, expectedB int) {

	qAPIArgs00 := []string{"example02", "query", "PEER0"}
	qAPIArgs01 := []string{"example02", "query", "PEER1"}
	qAPIArgs02 := []string{"example02", "query", "PEER2"}
	qAPIArgs03 := []string{"example02", "query", "PEER3"}
	qArgsa := []string{"a"}
	qArgsb := []string{"b"}

	res0A, _ := chaincode.QueryOnHost(qAPIArgs00, qArgsa)
	res0B, _ := chaincode.QueryOnHost(qAPIArgs00, qArgsb)
	res0AI, _ := strconv.Atoi(res0A)
	res0BI, _ := strconv.Atoi(res0B)

	res1A, _ := chaincode.QueryOnHost(qAPIArgs01, qArgsa)
	res1B, _ := chaincode.QueryOnHost(qAPIArgs01, qArgsb)
	res1AI, _ := strconv.Atoi(res1A)
	res1BI, _ := strconv.Atoi(res1B)

	res2A, _ := chaincode.QueryOnHost(qAPIArgs02, qArgsa)
	res2B, _ := chaincode.QueryOnHost(qAPIArgs02, qArgsb)
	res2AI, _ := strconv.Atoi(res2A)
	res2BI, _ := strconv.Atoi(res2B)

	res3A, _ := chaincode.QueryOnHost(qAPIArgs03, qArgsa)
	res3B, _ := chaincode.QueryOnHost(qAPIArgs03, qArgsb)
	res3AI, _ := strconv.Atoi(res3A)
	res3BI, _ := strconv.Atoi(res3B)

	fmt.Println("Results in a and b vp0 : ", res0AI, res0BI)
	fmt.Println("Results in a and b vp1 : ", res1AI, res1BI)
	fmt.Println("Results in a and b vp2 : ", res2AI, res2BI)
	fmt.Println("Results in a and b vp3 : ", res3AI, res3BI)

	if res0AI == expectedA && res1AI == expectedA && res2AI == expectedA && res3AI == expectedA {
		myStr := fmt.Sprintf("\n\n%s TEST PASS : Results in A value match on all Peers after %s", txName, txName)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
		myStr = fmt.Sprintf("Values Verified : peer0: %d, peer1: %d, peer2: %d, peer3: %d", res0AI, res1AI, res2AI, res3AI)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
	} else {
		overallTestPass = false
		myStr := fmt.Sprintf("\n\n%s TEST FAIL: Results in A value DO NOT match on all Peers after %s", txName, txName)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
	}

	if res0BI == expectedB && res1BI == expectedB && res2BI == expectedB && res3BI == expectedB {
		myStr := fmt.Sprintf("\n\n%s TEST PASS : Results in B value match on all Peers after %s", txName, txName)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
		myStr = fmt.Sprintf("Values Verified : peer0: %d, peer1: %d, peer2: %d, peer3: %d\n\n", res0BI, res1BI, res2BI, res3BI)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
	} else {
		overallTestPass = false
		myStr := fmt.Sprintf("\n\n%s TEST FAIL: Results in B value DO NOT match on all Peers after %s", txName, txName)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
	}
}

func getHeight_deprecated(expected int) {

	ht0, _ := chaincode.GetChainHeight("PEER0")
	ht1, _ := chaincode.GetChainHeight("PEER1")
	ht2, _ := chaincode.GetChainHeight("PEER2")
	ht3, _ := chaincode.GetChainHeight("PEER3")

	if (ht0 == expected) && (ht1 == expected) && (ht2 == expected) && (ht3 == expected) {
		myStr := fmt.Sprintf("CHAIN HEIGHT TEST PASS : Results in A value match on all Peers after deploy and single invoke:\n")
		myStr += fmt.Sprintf("  Height Verified: ht0=%d, ht1=%d, ht2=%d, ht3=%d", ht0, ht1, ht2, ht3)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
		writer.Flush()
	} else {
		overallTestPass = false
		myStr := fmt.Sprintf("CHAIN HEIGHT TEST FAIL : value in chain height DOES NOT MATCH expected value %d ON ALL PEERS after deploy and single invoke:\n", expected)
		myStr += fmt.Sprintf("  All heights DO NOT MATCH expected value: ht0=%d, ht1=%d, ht2=%d, ht3=%d", ht0, ht1, ht2, ht3)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
		writer.Flush()
	}
}
func getHeight(expectedToMatch int) {

	ht0, _ := chaincode.GetChainHeight("PEER0")
	ht1, _ := chaincode.GetChainHeight("PEER1")
	ht2, _ := chaincode.GetChainHeight("PEER2")
	ht3, _ := chaincode.GetChainHeight("PEER3")

	numPeers := peernetwork.GetNumberOfPeers(myNetwork)
	if numPeers != 4 { fmt.Println(fmt.Sprintf("TEST FAILURE: TODO: Must fix code %d peers, rather than default=4 peers in network!!!", numPeers)) }
	// before declaring failure, we will first check if we at least have consensus, with enough nodes with the correct height
	agree := 1
	if (ht0 == ht1) { agree++ }
	if (ht0 == ht2) { agree++ }
	if (ht0 == ht3) { agree++ }
	if agree < 3 {
		agree = 1
		if (ht1 == ht2) { agree++ }
		if (ht1 == ht3) { agree++ }
	}

	if (ht0 == expectedToMatch) && (ht1 == expectedToMatch) && (ht2 == expectedToMatch) && (ht3 == expectedToMatch) {
		myStr := fmt.Sprintf("CHAIN HEIGHT TEST PASS : value match on all Peers, after deploy and single invoke:\n")
		myStr += fmt.Sprintf("  Height Verified: ht0=%d, ht1=%d, ht2=%d, ht3=%d", ht0, ht1, ht2, ht3)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
		writer.Flush()
	} else if agree >= 3 {
		myStr := fmt.Sprintf("CHAIN HEIGHT TEST PASS : value match on enough Peers for consensus, after deploy and single invoke:\n")
		myStr += fmt.Sprintf("  Height Verified: ht0=%d, ht1=%d, ht2=%d, ht3=%d", ht0, ht1, ht2, ht3)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
		writer.Flush()
	} else {
		overallTestPass = false
		myStr := fmt.Sprintf("CHAIN HEIGHT TEST FAIL : value in chain height DOES NOT MATCH expected value %d ON ALL PEERS after deploy and single invoke:\n", expectedToMatch)
		myStr += fmt.Sprintf("  All heights DO NOT MATCH expected value: ht0=%d, ht1=%d, ht2=%d, ht3=%d", ht0, ht1, ht2, ht3)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
		writer.Flush()
	}
}

func getBlockTxInfo(blockNumber int) {
	errTransactions := 0
	height, _ := chaincode.GetChainHeight("PEER0")
	myStr := fmt.Sprintf("\n############### Total Blocks # %d\n", height)
	fmt.Printf(myStr)
	fmt.Fprintf(writer, myStr)

	for i := 1; i < height; i++ {
		//fmt.Printf("\n============================== Current BLOCKS %d ==========================\n", i)
		nonHashData, _ := chaincode.GetBlockTrxInfoByHost("PEER0", i)
		length := len(nonHashData.TransactionResult)
		for j := 0; j < length; j++ {
			// Print Error info only when transaction failed
			if nonHashData.TransactionResult[j].ErrorCode > 0 {
				myStr1 := fmt.Sprintln("\nBlock[%d] UUID [%d] ErrorCode [%d] Error: %s\n", i, nonHashData.TransactionResult[j].Uuid, nonHashData.TransactionResult[j].ErrorCode, nonHashData.TransactionResult[j].Error)
				fmt.Println(myStr1)
				fmt.Fprintln(writer, myStr1)
				errTransactions++
			}
		}
	}
	if errTransactions > 0 {
		myStr = fmt.Sprintf("\nTotal Blocks ERRORS # %d\n", errTransactions)
		fmt.Println(myStr)
		fmt.Fprintln(writer, myStr)
	}
	writer.Flush()
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	myStr := fmt.Sprintf("\n################# %s took %s \n", name, elapsed)
	fmt.Println(myStr)
	fmt.Fprintln(writer, myStr)
	myStr = fmt.Sprintf("################# Execution Completed #################")
	fmt.Fprintln(writer, myStr)
	fmt.Println(myStr)
	writer.Flush()
	myStr = fmt.Sprintf("\n################# %s took %s \n", name, elapsed)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}