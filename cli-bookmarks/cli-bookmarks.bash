# This file is part of cli-bookmarks.
#
# Copyright (C) 2018  David Gamba Rios
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

function cb() {
  local out=""
  local exit_value=1
  if [[ $# -eq 0 ]]; then
    out=`cli-bookmarks`
    exit_value=$?
  else
    out=`cli-bookmarks "$*"`
    exit_value=$?
  fi
  if [[ $exit_value == 0 ]]; then
    cd "$out"
  else
    echo "$out"
  fi
}

function _cliBookmarks() {
  COMPREPLY=(`cli-bookmarks --completion-current ${2} --completion-previous ${3}`)
  return 0
}
complete -o filenames -o nospace -F _cliBookmarks cb
