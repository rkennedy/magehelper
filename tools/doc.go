// Package tools offers Mage dependencies and commands to help install tools for use during a build, such as linters and
// code generators.
//
// The package is predicated on the principle that the tools that a project uses to build are considered part of the
// project. The project should track which versions of tools it uses, and the build should ensure that those versions of
// the tools are what get invoked.
package tools
