## 💡 What is this project?

Instead of running heavy Docker containers that eat up memory and take seconds to start, this project runs **tiny WebAssembly microservices directly on Flatcar Linux**.

By using WebAssembly instead of Docker, we get:

* ⚡ **Crazy Speed:** Handles over **119,000 real orders per second** with a response time of **0.6 milliseconds**.
* 💾 **Tiny Memory Footprint:** Each service runs in just **~12 MB of RAM** (instead of 120MB+ for Docker).
* 👥 **Extreme Density:** You can run **100 separate microservices on a single cheap $5 server (1GB RAM)** without it breaking a sweat or running out of memory.
