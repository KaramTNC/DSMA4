Description on how to use program:
The program is designed to work with exactly 3 clients.
The program can be started by running the go file Clients/Client.go. Please open the 3 Clients in 3 different terminals

When opening the program, the user will be promted to enter a unique Id for this node. Here the only valid options are 1, 2 or 3. 

After entering a id, the user will be asked if they want to enter the critical section. Please only enter the critical section, when all 3 clients are online.

After the client has entered the critical section, they will automaticly leave after 5 seconds. After leaving, other clients that have asked for the critical section will be granted access and the user will again again be promted to enter the critical section. 