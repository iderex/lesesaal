// Package web is the surface: the pages a volunteer classifies on, the pages a
// campaign owner reads results from, and the handlers behind both, served from
// the same process as everything else.
//
// It may read the core and the model interface. It may not read the store,
// because a handler that reaches the store directly is a rule about campaigns
// written where nothing tests it without a browser.
//
// There is no code in this package yet. #43 builds the classification page and
// #44 constrains what it may load.
package web
