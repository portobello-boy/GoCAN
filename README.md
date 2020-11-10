# GoCAN
An implementation of a distributed Content Addressable Network using Golang.

## Background
A content addressable network (CAN) is a collection of systems acting as a distributed database where data is addressed based on content, instead of location. This implementation is inspired by [this document](https://people.eecs.berkeley.edu/~sylvia/papers/cans.pdf), which details a method for distributing data across systems using a _d_-dimensional coordinate space. Myself and [TheBigFish](https://github.com/TheBiggerFish) implemented a version of this in C++ using our own packet types during a network theory class in University, so this will be a more refined approach.

## Resources
https://people.eecs.berkeley.edu/~sylvia/papers/cans.pdf

https://gobyexample.com/