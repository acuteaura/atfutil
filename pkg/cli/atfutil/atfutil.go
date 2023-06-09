package atfutil

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"

	"github.com/pkg/errors"

	"ops-networking/pkg/render"

	"ops-networking/pkg/netcalc"

	"github.com/go-yaml/yaml"
	"ops-networking/pkg/atf"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "atfutil",
	Short: "atfutil can validate and render atf (allocation table format) yaml files",
}

func quitWithError(err error) {
	fmt.Fprintf(os.Stderr, "err: %s\n", err.Error())
	os.Exit(1)
}

func getInputFile(inputFilename string) (*os.File, error) {
	var inputFile *os.File
	if inputFilename == "" {
		return nil, errors.New("need an input filename")
	}
	if inputFilename == "-" {
		inputFile = os.Stdin
	} else {
		file, err := os.OpenFile(inputFilename, os.O_RDONLY, 0)
		if err != nil {
			return nil, err
		}
		inputFile = file
	}
	return inputFile, nil
}

func getOutputFile(outputFilename string) (*os.File, error) {
	var outputFile *os.File
	if outputFilename == "" {
		return nil, errors.New("need an output filename")
	}
	if outputFilename == "-" {
		outputFile = os.Stdout
	} else {
		file, err := os.OpenFile(outputFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return nil, err
		}
		outputFile = file
	}
	return outputFile, nil
}

func loadAtfFromFile(inputFile *os.File) (*atf.File, error) {
	data, err := ioutil.ReadAll(inputFile)
	if err != nil {
		return nil, err
	}
	atf := new(atf.File)
	err = yaml.Unmarshal(data, atf)
	if err != nil {
		return nil, err
	}
	if atf.Superblock == nil {
		return nil, errors.New("file missing superblock")
	}
	for i, alloc := range atf.Allocations {
		if alloc.Network == nil {
			return nil, errors.Errorf("file missing network in allocation [%d]", i)
		}
	}
	return atf, nil
}

func netpoolFromAtf(atfFile *atf.File) (*netcalc.IPNetPool, error) {
	ipNet := make([]*net.IPNet, 0, len(atfFile.Allocations))
	for _, alloc := range atfFile.Allocations {
		ipNet = append(ipNet, alloc.Network.IPNet)
	}
	return netcalc.NewIPNetPool(atfFile.Superblock.String(), ipNet...)
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "validate an input file to be valid atf and have no network overlap",
	Run: func(cmd *cobra.Command, args []string) {
		inFile, err := getInputFile(*inputFilename)
		if err != nil {
			quitWithError(err)
		}
		defer inFile.Close()
		atf, err := loadAtfFromFile(inFile)
		if err != nil {
			quitWithError(err)
		}
		_, err = netpoolFromAtf(atf)
		if err != nil {
			quitWithError(err)
		}
		os.Exit(0)
	},
}

var renderCmd = &cobra.Command{
	Use:   "render",
	Short: "render an atf.yaml to a human readable format",
	Run: func(cmd *cobra.Command, args []string) {
		outBuffer := &bytes.Buffer{}

		inFile, err := getInputFile(*inputFilename)
		if err != nil {
			quitWithError(err)
		}
		defer inFile.Close()

		atfFile, err := loadAtfFromFile(inFile)
		if err != nil {
			quitWithError(err)
		}
		pool, err := netpoolFromAtf(atfFile)
		if err != nil {
			quitWithError(err)
		}

		switch *renderFormat {
		case "markdown":
			render.RenderPoolToMarkdown(outBuffer, atfFile, pool, *renderFree)
		default:
			quitWithError(errors.New("unknown render format"))
		}

		outFile, err := getOutputFile(*outputFilename)
		if err != nil {
			quitWithError(err)
		}
		defer outFile.Close()
		_, err = io.Copy(outFile, outBuffer)
		if err != nil {
			quitWithError(err)
		}

		os.Exit(0)
	},
}

var allocCmd = &cobra.Command{
	Use:   "alloc",
	Short: "allocate a new subnet",
	Long:  "allocate a new subnet, the smallest fitting free slice is automatically found and allocated to keep your IP space fragmentation low",
	Run: func(cmd *cobra.Command, args []string) {
		inFile, err := getInputFile(*inputFilename)
		if err != nil {
			quitWithError(err)
		}
		defer inFile.Close()

		atfFile, err := loadAtfFromFile(inFile)
		if err != nil {
			quitWithError(err)
		}
		pool, err := netpoolFromAtf(atfFile)
		if err != nil {
			quitWithError(err)
		}

		superAllocSize, _ := atfFile.Superblock.Mask.Size()

		if *allocSize > netcalc.AWS_MIN_SUBNET_SIZE || *allocSize <= superAllocSize {
			quitWithError(errors.Errorf("requested block size is out of range (%d < block < %d)", superAllocSize, netcalc.AWS_MIN_SUBNET_SIZE))
		}

		net, err := pool.Alloc(*allocSize)
		if err != nil {
			quitWithError(err)
		}

		atfFile.Allocations = append(atfFile.Allocations, atf.Allocation{
			Network:     &atf.IPNet{net},
			Description: *allocDesc,
		})

		outBytes, err := yaml.Marshal(atfFile)
		if err != nil {
			quitWithError(err)
		}

		outFile, err := getOutputFile(*outputFilename)
		if err != nil {
			quitWithError(err)
		}
		defer outFile.Close()
		outFile.Write(outBytes)
		if err != nil {
			quitWithError(err)
		}

		os.Exit(0)
	},
}

func Command() *cobra.Command {
	return rootCmd
}

var inputFilename *string
var outputFilename *string

var renderFree *bool
var renderFormat *string

var allocSize *int
var allocDesc *string

func init() {
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(renderCmd)
	rootCmd.AddCommand(allocCmd)

	inputFilename = rootCmd.PersistentFlags().StringP("input-file", "i", "-", "input file")
	outputFilename = rootCmd.PersistentFlags().StringP("output-file", "o", "-", "output file")

	renderFree = renderCmd.Flags().BoolP("all-blocks", "a", false, "include free blocks when rendering")
	renderFormat = renderCmd.Flags().StringP("render-format", "f", "markdown", "render format (markdown)")

	allocSize = allocCmd.Flags().IntP("size", "s", -1, "size of the network to allocate")
	allocDesc = allocCmd.Flags().StringP("description", "d", "", "description for the newly allocated subnet")
}
