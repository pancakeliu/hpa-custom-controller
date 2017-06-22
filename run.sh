#!/bin/bash

cd /root
./hpa-custom-controller --leader-election-id=$(hostname)
