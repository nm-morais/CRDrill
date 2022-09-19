install:
	go build -o CRDrill .
	sudo chmod +x CRDrill
	sudo mv CRDrill /usr/local/bin
