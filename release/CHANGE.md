# v0.1.0
We have conducted comprehensive testing on all major functionalities, with the primary test configurations located in the `testconfig` directory. Error messaging within the code has also been significantly enhanced—for instance, different providers (Cloudflare, Tencent, Aliyun, etc.) now surface 
explicit error indications when upstream requests return errors.
Additionally, we have added support for Lines (routing policies) for both Aliyun and Tencent Cloud. As a result, 
LightDDNS now enables domain-name resolution with per-carrier/operator differentiation (effectively handcrafting BGP—nice!). 
Furthermore, other domain settings, such as TTL, can now be updated in the provider’s data in accordance with our configuration changes, 
thanks to our new diff model.

There are many more updates—please refer to the file changes within the commits for full details.

# v0.0.0-alpha.3
In this version, we mainly optimized the build process. 
We no longer upload extra lightddns files, and we've also eliminated build.json in favor of using flags to pass data.

which simplifies the process in scripts or shell environments.

# v0.0.0-alpha.2
In this release, we've completed basic testing and packaging for different Linux distributions.

We've also written a Github CI script for automated deployment.

The project is stabilizing but requires further testing. The current version is v0.0.1 (API), 
and specific behavior, such as configuration files, may change in the future.

# v0.0.0-alpha.1
First Release Version
