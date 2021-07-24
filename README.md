<p align="center">
  <img src="./ntgcalls.png" alt="ntgcalls" />
</p>

# NativeTgCalls
This Go project lets you work with Telegram calls using Pion mediadevices.

## Why Go instead of C++ or Node.js?
#### 1) It is a simpler low-level language than C++.
#### 2) It can get compiled to assembly, which makes it much faster than interpreted JavaSript.
#### 3) It is cross-platform and uses less resources.

## Getting started
Compile the Go code:
```bash
cd dist/
go build -o ntgcalls.so -buildmode=c-shared .
```
Specify the client configuration, add the chat ID in test.py and run it:
```bash
python test.py
```

## What type of input is needed?
The same type of the stable [pytgcalls].

## Is it really working?
Not really, but any type of [help is accepted].

[pytgcalls]: https://github.com/pytgcalls/pytgcalls
[help is accepted]: https://github.com/pion/mediadevices/issues/339
