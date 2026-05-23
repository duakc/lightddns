# lightddns
Light weight DDNS Prog


## TODO
- [ ] Add Prometheus (Open Telemetry)
- [ ] Add an observed logging system
- [ ] Complete doc
- [ ] Web config generator (in the docs page)
- [ ] Add more providers ( aws, tencent, alibaba ... )

## Considering TODOs
Below Todos need more consideration.

- [ ] Add Notify ( slack, telegram, ... )

For this project, I am striving to integrate comprehensive metric data while adhering as closely as possible to the existing feature set. 
I believe that, given the use of Prometheus for data processing, notification capabilities are already achievable; 
therefore, there is no need for me to implement a separate notification system.

However, implementing this functionality internally would allow for the exposure of various internal variables and metrics, 
while also facilitating better overall management.