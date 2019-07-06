# Versioning
[Semantic versioning](https://semver.org/) simply doesn't work in the real world and provide no value. 
Major versions promote releasing software with breaking changes, i.e., is not backward compatible. By definition,
a major version change is saying if you upgrade to this version it will (may) break your software. As professional
software developers, we never want to release a new version of software that will breake consumers of that software.
If we indeed create a new software version that is not designed for the existing consuber base, then we have created
a new and different piece of software. The best way to communicate that is by simply giving the software a new name.
In go, this means importing a different module - done.

**Never make any breaking changes to a software module**. You can always add more to a module, but you can
never remove. 

## Is foo v2.1.0 older or newer then bar v5.3.2 ?
Another problem with the classic versioning scheme is that the version does not convey any information of how old the
software is. For example, v5.3.2 might have been released 11 years ago, but v2.1.0 was released yesterday. Instead, get
some value out of the version by embedding the release date, e.g., **v2019.0706.151102** - which says the sofware version 
was released on 2019-07-06T15:11:02

