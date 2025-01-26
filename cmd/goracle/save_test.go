package main

// func TestSave(t *testing.T) {

// 	check := assert.New(t)

// 	cli := SetupCLI(t)
// 	cli.Env.Args = []string{"goracle", "info", "--priv", "../../testdata/george.pem"}

// 	ctx := context.TODO()
// 	//cli.Run(ctx)

// 	exe := cli.Obj()
// 	george := exe.Self

// 	ringo := new(gork.Principal)
// 	ringoPem, err := cli.Env.Filesystem.Open("../../testdata/george.pem")
// 	check.NoError(err)

// 	err = ringo.FromPem(ringoPem)
// 	check.NoError(err)

// 	george.AddPeer(ringo.AsPeer())

// 	exe.Save(ctx, )

// }
