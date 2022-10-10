# go for bitcoin network

Several years ago I went into the bitcoin source code. I made some modifications in a node. I opened the numbers of maximum peers from 30 to more than 200. Also, all the transactions, and all other peers connected, were sent to an on-line database.
I sent dozens of these probes, as part of the bitcoin net. This allowed me have a good estimation from which country a transaction was sent : it was where I saw them at first. Also I could show some geographical statistics in real time.

The original code was on the repository of the firm onBlock, which was private.

Except the modified bitcoin probes, everything was written in golang.

Here are various golang source code for this project. I do not have everything to make it work again (database, servers).
I used postgres database, websocket server and client, and displays realtime information on a graphic webpage (with google jsapi)

The program testnet.go might just look interesting as at the beginning it displays all the network interfaces and status of its host.
