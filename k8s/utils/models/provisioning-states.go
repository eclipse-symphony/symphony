/*

	MIT License

	Copyright (c) Microsoft Corporation.

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE

*/

package provisioningstates

// ARM has a strict requirement on what the terminal states need to be for the provisioning of resources through the LRO contract (Succeeded, Failed, Cancelled)
// The documentation that talks about this can be found here: https://armwiki.azurewebsites.net/rpaas/async.html#provisioningstate-property
// The below exported members capture these states. The first three are the terminal states required by ARM and the
// fourth is a non-terminal state we use to indicate that the resource is being reconciled.
const (
	Succeeded   = "Succeeded"
	Failed      = "Failed"
	Cancelled   = "Cancelled"
	Reconciling = "Reconciling"
)
