#!/bin/sh
#
# md5str.sh
#
# Example command-line executable to process an input argument
#
# @author      Nicola Asuni <info@tecnick.com>
# ------------------------------------------------------------------------------

echo -n $1 | md5sum | awk '{print $1}'
