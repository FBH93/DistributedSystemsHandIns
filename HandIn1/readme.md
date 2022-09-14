Run the solution for the Dining Philosophers problem using the command 

`go run HandIn1.go`

in a command prompt in the HandIn1-folder.

# Output

`Starting Program...` is printed as the program begins execution

`Finish program` is printed as the program ends (and thus, all child routines are killed)

The program prints events as they occur, for instance:

`Philosopher y is thinking...`

`Philosopher y is eating with fork x and z for the n time`

`Philosopher y is thinking again with thought i...` 

Note that the order of prints to the command prompt might differ from its underlying operations due to interleaving between routines.

## Example output

```
Philosopher 1 is thinking...
Philosopher 5 is thinking...
Philosopher 4 is thinking...
Philosopher 2 is thinking...
Philosopher 3 is thinking...
Philosopher 2 is eating with fork 1 and 2 for the 1 time
Philosopher 2 is thinking again with thought 1...
Philosopher 5 is eating with fork 4 and 5 for the 1 time
Philosopher 5 is thinking again with thought 1...
Philosopher 1 is eating with fork 5 and 1 for the 1 time
Philosopher 1 is thinking again with thought 1...
Philosopher 2 is eating with fork 1 and 2 for the 2 time
Philosopher 2 is thinking again with thought 2...
Philosopher 5 is eating with fork 4 and 5 for the 2 time
Philosopher 5 is thinking again with thought 2...
Philosopher 1 is eating with fork 5 and 1 for the 2 time
Philosopher 1 is thinking again with thought 2...
Philosopher 1 is eating with fork 5 and 1 for the 3 time
Philosopher 3 is eating with fork 3 and 2 for the 1 time
Philosopher 1 is thinking again with thought 3...
##### Philosopher 1 is finished eating #####
Philosopher 4 is eating with fork 3 and 4 for the 1 time
Philosopher 4 is thinking again with thought 1...
Philosopher 3 is thinking again with thought 1...
Philosopher 3 is eating with fork 3 and 2 for the 2 time
Philosopher 3 is thinking again with thought 2...
Philosopher 5 is eating with fork 4 and 5 for the 3 time
Philosopher 5 is thinking again with thought 3...
##### Philosopher 5 is finished eating #####
Philosopher 2 is eating with fork 1 and 2 for the 3 time
Philosopher 2 is thinking again with thought 3...
##### Philosopher 2 is finished eating #####
Philosopher 4 is eating with fork 3 and 4 for the 2 time
Philosopher 4 is thinking again with thought 2...
Philosopher 4 is eating with fork 3 and 4 for the 3 time
Philosopher 4 is thinking again with thought 3...
##### Philosopher 4 is finished eating #####
Philosopher 3 is eating with fork 3 and 2 for the 3 time
Philosopher 3 is thinking again with thought 3...
##### Philosopher 3 is finished eating #####
Finish program
```



