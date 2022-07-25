// Copyright (c) 2012-2018 Ugorji Nwoke. All rights reserved.
// Use of this source code is governed by a MIT license found in the LICENSE file.

//go:build alltests && go1.7
// +build alltests,go1.7

package codec

// Run this using:
//   go test -tags=alltests -run=Suite -coverprofile=cov.out
//   go tool cover -html=cov.out
//
// Because build tags are a build time parameter, we will have to test out the
// different tags separately.
// Tags: x codecgen safe appengine notfastpath
//
// These tags should be added to alltests, e.g.
//   go test '-tags=alltests x codecgen' -run=Suite -coverprofile=cov.out
//
// To run all tests before submitting code, run:
//    a=( "" "safe" "codecgen" "notfastpath" "codecgen notfastpath" "codecgen safe" "safe notfastpath" )
//    for i in "${a[@]}"; do echo ">>>> TAGS: $i"; go test "-tags=alltests $i" -run=Suite; done
//
// This only works on go1.7 and above. This is when subtests and suites were supported.

import "testing"

// func TestMain(m *testing.M) {
// 	println("calling TestMain")
// 	// set some parameters
// 	exitcode := m.Run()
// 	os.Exit(exitcode)
// }

func testGroupResetFlags() {
	testUseMust = false
	testCanonical = false
	testUseMust = false
	testInternStr = false
	testUseIoEncDec = -1
	testStructToArray = false
	testCheckCircRef = false
	testUseReset = false
	testMaxInitLen = 0
	testUseIoWrapper = false
	testNumRepeatString = 8
	testEncodeOptions.RecursiveEmptyCheck = false
	testDecodeOptions.MapValueReset = false
	testUseIoEncDec = -1
	testDepth = 0
}

func testSuite(t *testing.T, f func(t *testing.T)) {
	// find . -name "*_test.go" | xargs grep -e 'flag.' | cut -d '&' -f 2 | cut -d ',' -f 1 | grep -e '^test'
	// Disregard the following: testInitDebug, testSkipIntf, testJsonIndent (Need a test for it)

	testReinit() // so flag.Parse() is called first, and never called again

	testDecodeOptions = DecodeOptions{}
	testEncodeOptions = EncodeOptions{}

	testGroupResetFlags()

	testReinit()
	t.Run("optionsFalse", f)

	testCanonical = true
	testUseMust = true
	testInternStr = true
	testUseIoEncDec = 0
	// xdebugf("setting StructToArray=true")
	testStructToArray = true
	testCheckCircRef = true
	testUseReset = true
	testDecodeOptions.MapValueReset = true
	testEncodeOptions.RecursiveEmptyCheck = true
	testReinit()
	t.Run("optionsTrue", f)

	// xdebugf("setting StructToArray=false")
	testStructToArray = false
	testDepth = 6
	testReinit()
	t.Run("optionsTrue-deepstruct", f)
	testDepth = 0

	// testEncodeOptions.AsSymbols = AsSymbolAll
	testUseIoWrapper = true
	testReinit()
	t.Run("optionsTrue-ioWrapper", f)

	testUseIoEncDec = -1

	// make buffer small enough so that we have to re-fill multiple times.
	testSkipRPCTests = true
	testUseIoEncDec = 128
	// testDecodeOptions.ReaderBufferSize = 128
	// testEncodeOptions.WriterBufferSize = 128
	testReinit()
	t.Run("optionsTrue-bufio", f)
	// testDecodeOptions.ReaderBufferSize = 0
	// testEncodeOptions.WriterBufferSize = 0
	testUseIoEncDec = -1
	testSkipRPCTests = false

	testNumRepeatString = 32
	testReinit()
	t.Run("optionsTrue-largestrings", f)

	// The following here MUST be tested individually, as they create
	// side effects i.e. the decoded value is different.
	// testDecodeOptions.MapValueReset = true // ok - no side effects
	// testDecodeOptions.InterfaceReset = true // error??? because we do deepEquals to verify
	// testDecodeOptions.ErrorIfNoField = true // error, as expected, as fields not there
	// testDecodeOptions.ErrorIfNoArrayExpand = true // no error, but no error case either
	// testDecodeOptions.PreferArrayOverSlice = true // error??? because slice != array.
	// .... however, update deepEqual to take this option
	// testReinit()
	// t.Run("optionsTrue-resetOptions", f)

	testGroupResetFlags()
}

/*
find . -name "codec_test.go" | xargs grep -e '^func Test' | \
    cut -d '(' -f 1 | cut -d ' ' -f 2 | \
    while read f; do echo "t.Run(\"$f\", $f)"; done
*/

func testCodecGroup(t *testing.T) {
	// println("running testcodecsuite")
	// <setup code>

	testMsgpackGroup(t)
	// testSimpleMammothGroup(t)
	// testRpcGroup(t)
	testNonHandlesGroup(t)

	// <tear-down code>
}

func testMsgpackGroup(t *testing.T) {
	t.Run("TestMsgpackCodecsTable", TestMsgpackCodecsTable)
	t.Run("TestMsgpackCodecsMisc", TestMsgpackCodecsMisc)
	t.Run("TestMsgpackCodecsEmbeddedPointer", TestMsgpackCodecsEmbeddedPointer)
	t.Run("TestMsgpackStdEncIntf", TestMsgpackStdEncIntf)
	t.Run("TestMsgpackMammoth", TestMsgpackMammoth)
	t.Run("TestMsgpackRaw", TestMsgpackRaw)
	t.Run("TestMsgpackRpcGo", TestMsgpackRpcGo)
	t.Run("TestMsgpackRpcSpec", TestMsgpackRpcSpec)
	t.Run("TestMsgpackSwallowAndZero", TestMsgpackSwallowAndZero)
	t.Run("TestMsgpackRawExt", TestMsgpackRawExt)
	t.Run("TestMsgpackMapStructKey", TestMsgpackMapStructKey)
	t.Run("TestMsgpackDecodeNilMapValue", TestMsgpackDecodeNilMapValue)
	t.Run("TestMsgpackEmbeddedFieldPrecedence", TestMsgpackEmbeddedFieldPrecedence)
	t.Run("TestMsgpackLargeContainerLen", TestMsgpackLargeContainerLen)
	t.Run("TestMsgpackMammothMapsAndSlices", TestMsgpackMammothMapsAndSlices)
	t.Run("TestMsgpackTime", TestMsgpackTime)
	t.Run("TestMsgpackUintToInt", TestMsgpackUintToInt)
	t.Run("TestMsgpackDifferentMapOrSliceType", TestMsgpackDifferentMapOrSliceType)
	t.Run("TestMsgpackScalars", TestMsgpackScalars)
	t.Run("TestMsgpackOmitempty", TestMsgpackOmitempty)
	t.Run("TestMsgpackIntfMapping", TestMsgpackIntfMapping)
	t.Run("TestMsgpackMissingFields", TestMsgpackMissingFields)
	t.Run("TestMsgpackMaxDepth", TestMsgpackMaxDepth)

	t.Run("TestMsgpackDecodeMapAndExtSizeMismatch", TestMsgpackDecodeMapAndExtSizeMismatch)
}

func testRpcGroup(t *testing.T) {
	t.Run("TestMsgpackRpcGo", TestMsgpackRpcGo)
	t.Run("TestMsgpackRpcSpec", TestMsgpackRpcSpec)
}

func testNonHandlesGroup(t *testing.T) {
	// grep "func Test" codec_test.go | grep -v -E '(Cbor|Json|Simple|Msgpack|Binc)'
	//t.Run("TestBufioDecReader", TestBufioDecReader)
	//t.Run("TestAtomic", TestAtomic)
	//t.Run("TestAllEncCircularRef", TestAllEncCircularRef)
	//t.Run("TestAllAnonCycle", TestAllAnonCycle)
	//t.Run("TestMultipleEncDec", TestMultipleEncDec)
	//t.Run("TestAllErrWriter", TestAllErrWriter)
}

func TestCodecSuite(t *testing.T) {
	testSuite(t, testCodecGroup)

	testGroupResetFlags()

	testMaxInitLen = 10
	testReinit()
	oldWriteExt := testMsgpackH.WriteExt
	oldNoFixedNum := testMsgpackH.NoFixedNum

	testMsgpackH.WriteExt = !testMsgpackH.WriteExt
	testReinit()
	t.Run("msgpack-inverse-writeext", testMsgpackGroup)

	testMsgpackH.WriteExt = oldWriteExt

	testMsgpackH.NoFixedNum = !testMsgpackH.NoFixedNum
	testReinit()
	t.Run("msgpack-fixednum", testMsgpackGroup)

	testMsgpackH.NoFixedNum = oldNoFixedNum

	testReinit()

	oldRpcBufsize := testRpcBufsize
	testRpcBufsize = 0
	t.Run("rpc-buf-0", testRpcGroup)
	testRpcBufsize = 0
	t.Run("rpc-buf-00", testRpcGroup)
	testRpcBufsize = 0
	t.Run("rpc-buf-000", testRpcGroup)
	testRpcBufsize = 16
	t.Run("rpc-buf-16", testRpcGroup)
	testRpcBufsize = 2048
	t.Run("rpc-buf-2048", testRpcGroup)
	testRpcBufsize = oldRpcBufsize

	testGroupResetFlags()
}

// func TestCodecSuite(t *testing.T) {
// 	testReinit() // so flag.Parse() is called first, and never called again
// 	testDecodeOptions, testEncodeOptions = DecodeOptions{}, EncodeOptions{}
// 	testGroupResetFlags()
// 	testReinit()
// 	t.Run("optionsFalse", func(t *testing.T) {
// 		t.Run("TestJsonMammothMapsAndSlices", TestJsonMammothMapsAndSlices)
// 	})
// }
