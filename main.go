// Copyright 2025 Ivan Guerreschi. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import . "modernc.org/tk9.0"

func main() {
	Pack(Button(Txt("Hello"), Command(func() { Destroy(App) })))
	App.Wait()
}
