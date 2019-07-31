# goloadtest

An example of implementation of runnning test of huge number of task using compute power of different clusters(machines).Here we have made apiExecutor.go where we defined our task. All it does is make a simple get request to a specific url as specified in the .env file. We can run this task multiple times by running a slave cluster. We can have multile slave cluster which send the summary of task to master cluster. Each slave cluster can run number of tasks. We can run different slave clusters to make more task execute in differet machine. This can help to run a complex task in differentmachine so that more compute power can be achieved.

The message between master and slave are exchaged using grpc protocol. See result/result.proto for more details.
 
- Supports multiple clusters
- Push the summary of every cluster to the master cluster

## Build
-------
  1. create new .env file and paste the contents of .env_example
  2. run the build command
  ```
  make build
  ```


## Architecture
----------------

  - A slave is the cluster who send the summary of task to master
  - A master is the cluste who give command to slave and listens for task summary
  - The slave and master can be run in differetn machines as require

  * A single master cluster can be run by the following command

  ```
  ./goloadtest MASTER SERVER_IP PORT NAME_OF_MASTER
  
  eg : ./goloadtest MASTER 0.0.0.0 5001 testmaster
  ```
      
  Here MASTER is the name of the mode. The SERVER_IP and PORT is ip and port of the server to listen to.



  * A slave cluster can be run as 

  ```
  ./goloadtest SLAVE SERVER_IP PORT UNIQUE_NAME_OF_CLIENT NUM_OF_TASKS_TO_RUN BATCH_NUM_OF_TASKS_TO_RUN
  
  eg: ./goloadtest SLAVE 192.168.0.2 5001 testslave1 10000 1000
  ```

  Here the SLAVE is the name of mode. The SERVER_IP and PORT is the ip and port at which MASTER is listening at.

    * UNIQUE_NAME_OF_CLIENT     :  The unique client name. Each slave must have this param unique

    * NUM_OF_TASKS_TO_RUN       :  Total number of task to be run

    * BATCH_NUM_OF_TASKS_TO_RUN :  The tasks are run in batch. This param specifies how many task to run at once
    

## Running the app
--------------------

1. First run the master
2. Then run slaves
3. Once all slaves are connected type "start" in the master console. After this all the slaves must do their tasks

