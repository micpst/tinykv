#!/bin/bash

kill $(pgrep -f nginx) || true
