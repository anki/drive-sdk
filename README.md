Anki Drive SDK
==============

The Anki Drive SDK is a C implementation of the message protocols and data parsing routines necessary for communicating with Anki Drive vehicles. 

The Anki Drive SDK is licensed under the Apache 2.0 license. Projects in the `examples` folder are released under separate licenses.

- web: http://anki.github.io/drive-sdk/
- StackOverflow Tag: anki-drive-sdk
- Issues: [GitHub Issues](https://github.com/anki/drive-sdk/issues)
- Guides: [Getting Started on Ubuntu][ubuntu-getting-started], [Programming Guide][]
- Mailing list: [Anki-Dev Google Group](https://groups.google.com/a/anki.com/forum/#!forum/anki-dev)

[Programming Guide]: http://developer.anki.com/drive-sdk/docs/programming-guide


Features
========

The initial release of the SDK has the minimal number of functions required to use the message protocol and parse information sent by vehicles.

* Parse vehicle info from Bluetooth LE `LOCAL_NAME` and `MANUFACTURER_DATA` advertisement data.
* Create messages to send commands to vehicles.
* Parse vehicle response messages.

The initial version provides a limited subset of the message protocol.
We hope to expand available messages and functionality in the future, once we can do so in a safe and stable way.

Optional dependencies
=====================

* Building the example programs for linux requires [glib2][] and [libreadline][].
  See the guide to [Getting Started on Ubuntu][ubuntu-getting-started] for more info.

[ubuntu-getting-started]: https://github.com/anki/drive-sdk/wiki/Getting-Started-on-Ubuntu
[glib2]: http://www.gtk.org/
[libreadline]: http://cnswww.cns.cwru.edu/php/chet/readline/rltop.html

Building the SDK
================

The Anki Drive SDK builds cleanly on Mac OS X and most linux distributions.
The SDK builds the `libankidrive` library, with public headers available in `include`.

The SDK is built using `CMake 2.8` (<http://www.cmake.org>) on all platforms.

     $ mkdir -p build && cd build

     # build & test library
     $ cmake ..
     $ make
     $ make test

     # build library + examples (examples currently require Linux)
     $ cmake .. -DBUILD_EXAMPLES=ON
     $ make
     $ make test
     $ make install
     # install creates a build/dist/bin folder containing executables


### Run Unit Tests

     $ cd build
     $ cmake ..
     $ make
     $ make test


