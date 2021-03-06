package main
/******************** Testing Objective consensu:STATE TRANSFER ********
*   Setup: 5 node local docker peer network with security
*   0. Deploy chaincodeexample02 with 100, 200 as initial args
*   1. PAUSE ONE PEER1
*   2. Send ONE INVOKE REQUEST
*   3. Unpause Paused PEER1
*   4. Do A Query ON same PEER0 and PEER1 as in step3
*   5. Verify query results match on PEER0 and PEER1 after invoke
*********************************************************************/


import (
	"fmt"
	"strconv"
	"time"

	"obcsdk/chaincode"
	"obcsdk/peernetwork"
)

func main() {

	var MyNetwork peernetwork.PeerNetwork

	fmt.Println("Creating a local docker network")
	peernetwork.SetupLocalNetwork(4, false)

	time.Sleep(60000 * time.Millisecond)
	peernetwork.PrintNetworkDetails()
	_ = chaincode.InitNetwork()
	chaincode.InitChainCodes()
	chaincode.RegisterUsers()


	//get a URL details to get info n chainstats/transactions/blocks etc.
	aPeer, _ := peernetwork.APeer(chaincode.ThisNetwork)
	url := "http://" + aPeer.PeerDetails["ip"] + ":" + aPeer.PeerDetails["port"]


	fmt.Println("Peers on network ")
	chaincode.NetworkPeers(url)


	fmt.Println("\nPOST/Chaincode: Deploying chaincode at the beginning ....")
	dAPIArgs0 := []string{"example02", "init"}
	depArgs0 := []string{"a", "100000", "b", "90000"}
	chaincode.Deploy(dAPIArgs0, depArgs0)

	var inita, initb, curra, currb, j  int
	inita = 100000
	initb = 90000
	curra = inita
	currb = initb

	time.Sleep(60000 * time.Millisecond);
	fmt.Println("\nPOST/Chaincode: Querying a and b after deploy >>>>>>>>>>> ")
	qAPIArgs0 := []string{"example02", "query"}
	qArgsa := []string{"a"}
	qArgsb := []string{"b"}
	A, _ := chaincode.Query(qAPIArgs0, qArgsa)
	B, _ := chaincode.Query(qAPIArgs0, qArgsb)
	myStr := fmt.Sprintf("\nA = %s B= %s", A,B)
	fmt.Println(myStr)


	fmt.Println("******************************")
	fmt.Println("STOPPING PEER1 To Test Consensus")
	fmt.Println("******************************")

	peersToStop := []string{"PEER1", "PEER2"}
	peernetwork.StopPeersLocal(MyNetwork, peersToStop)

	j = 0
	for j < 5 {
		iAPIArgs0 := []string{"example02", "invoke"}
		invArgs0 := []string{"a", "b", "1"}
		invRes, _ := chaincode.Invoke(iAPIArgs0, invArgs0)
		fmt.Println("\nFrom Invoke invRes ", invRes)
		curra = curra - 1
		currb = currb + 1
		time.Sleep(30000 * time.Millisecond)
		j++

	}
	fmt.Println("UNPAUSING PEER1, ... To Test Consensus STATE TRANSFER")
	peernetwork.StartPeerLocal(MyNetwork, "PEER1")
	fmt.Println("Sleeping for 2 minutes for PEER1 to sync up - state transfer")
	//fmt.Println("Sleeping for 2 minutes ")
	time.Sleep(180000 * time.Millisecond)

	fmt.Println("\nPOST/Chaincode: Querying a and b after invoke >>>>>>>>>>> ")
	qAPIArgs00 := []string{"example02", "query", "PEER0"}
	qAPIArgs01 := []string{"example02", "query", "PEER1"}
	//qArgsa = []string{"a"}
	//qArgsb = []string{"b"}

        fmt.Println("Querying on PEER0")
	res0A, _ := chaincode.QueryOnHost(qAPIArgs00, qArgsa)
	res0B, _ := chaincode.QueryOnHost(qAPIArgs00, qArgsb)

	res0AI, _ := strconv.Atoi(res0A)
	res0BI, _ := strconv.Atoi(res0B)

	time.Sleep(180000 * time.Millisecond)
        fmt.Println("Querying on PEER1")
	res1A, _ := chaincode.QueryOnHost(qAPIArgs01, qArgsa)
	res1B, _ := chaincode.QueryOnHost(qAPIArgs01, qArgsb)

	res1AI, _ := strconv.Atoi(res1A)
	res1BI, _ := strconv.Atoi(res1B)

	if (curra == res0AI) && (currb == res0BI) {
		fmt.Println("Results in a and b match with Invoke values on PEER0: TEST CASE PASS")
		valueStr := fmt.Sprintf(" curra : %d, currb : %d, resa : %d , resb : %d", curra, currb, res0AI, res0BI)
		fmt.Println(valueStr)
	} else {
		fmt.Println("******************************")
		fmt.Println("RESULTS DO NOT MATCH on PEER0 : TEST CASE FAIL")

		fmt.Println("******************************")
	}

	if (curra == res1AI) && (currb == res1BI) {
		fmt.Println("Results in a and b match with Invoke values on PEER1: TEST CASE PASS")
		valueStr := fmt.Sprintf(" curra : %d, currb : %d, resa : %d , resb : %d", curra, currb, res1AI, res1BI)
		fmt.Println(valueStr)
	} else {
		fmt.Println("******************************")
		fmt.Println("RESULTS DO NOT MATCH on PEER1 : TEST CASE FAIL")
		fmt.Println("******************************")
	}

	j = 0
	for j < 5 {
		iAPIArgs0 := []string{"example02", "invoke"}
		invArgs0 := []string{"a", "b", "1"}
		invRes, _ := chaincode.Invoke(iAPIArgs0, invArgs0)
		fmt.Println("\nFrom Invoke invRes ", invRes)
		curra = curra - 1
		currb = currb + 1
		time.Sleep(30000 * time.Millisecond)
		j++
	}
  ht1, _ := chaincode.GetChainHeight("PEER1")
	fmt.Println("Ht on PEER1", ht1)

	ht0, _ := chaincode.GetChainHeight("PEER0")
	fmt.Println("Ht on PEER0", ht0)

	}
