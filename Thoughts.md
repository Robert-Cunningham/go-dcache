We want to emulate a single threaded dcache, except optimize with multiple threads the areas which can be optimized. That means that we should never reject requests. When we receive a request, we should check to see if it is blocked by any currently running threads. 

We will store requests in a staging area, while we batch them together and assemble a DAG. Then, we run a topological sort, which will transparently reorder the requests, and execute all of them. 

Making any guarantees about which instructions execute in what order is fruitless, since we could call Execute() at any time. So, why don't we make the Execute() call explicit, and guarantee that dependencies within the system at that point will execute equivalently to the order in which they were recieved.

Alternatively, we could just immediately run instructions which we recieve, but wait on them if something which blocks is currently running or waiting. This seems like a better system. 

We guarantee that instructions will run in the order in which they are recieved by the dcache system (or at least transparently relative to this). 