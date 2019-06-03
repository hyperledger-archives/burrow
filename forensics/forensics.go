// Copyright 2019 Monax Industries Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build forensics

// This package contains tools for examining, replaying, and debugging Tendermint-side and Burrow-side blockchain state.
// Some code is quick and dirty from particular investigations and some is better extracted, encapsulated and generalised.
// The sketchy code is included so that useful tools can be progressively put together as the generality of the types of
// forensic debugging needed in the wild are determined.

package forensics
