#!/usr/bin/env bash

## This script just takes any number of input files, and outputs them to a location (default pwd) using given flags
## The input/output file flags are set as follows:
## -f input_1.jpg -o input_1.txt [-c input_1_scaled.jpg]

flags=""
out_path=""
scaled_path=""
save_scaled=false

Usage() {
    cat <<-__EOF_
${0##*/} [-rnA -h height -w width -s scale -o output path -c scaled path] input files...
    -r         : Overwrite existing files
    -n         : ASCII as negative
    -A         : Print ASCII output
    -o path    : Output path for ASCII/text files
    -c path    : Path to save scaled images to
    -h int     : Height
    -w int     : Width
    -s float   : Scale factor
    -H         : Display this help message

To ASCIIfy a bunch of files in a given directory at half-size, and have the scaled copies stored in a different directory
and save the ASCII files in yet another, the command would look something like this:

    ${0##*/} -o ascii_output_dir -c scaled_img_dir -s 0.5 path/to/input/*.jpg
    ## Doing the same but scale all images to a uniform size:
    ${0##*/} -o ascii_output_dir -c scaled_img_dir -w 100 -h 150 path/to/input/*.jpg

__EOF_
}

while getopts :rns:w:h:Ao:c: f; do
    case $f in
        r|n|A)
            flags="${flags} -${f}"
            ;;
        s|w|h)
            flags="${flags} -${f} ${OPTARG}"
            ;;
        H)
            Usage
            exit 0
            ;;
        o)
            out_path="${OPTARG}/"
            ;;
        c)
            scaled_path="${OPTARG}/"
            save_scaled=true
            ;;
        *)
            Usage
            exit 1
            ;;
    esac
    ## Remove flag from input, we should be left with nothing but the files
done
shift $((OPTIND - 1 ))

## The loop:
for in_file in "${@}"; do
    ## input file excluding the extension
    name="${in_file%%.*}"
    ## also remove any leading path stuff
    name="${name##*/}"
    if $save_scaled ; then
        ## Get the file extension
        ext="${in_file##*.}"
        asciify $flags -f "${in_file}" -o "${out_path}${name}.txt" -c "${scaled_path}${name}_scaled.${ext}"
    else
        asciify $flags -f "${in_file}" -o "${out_path}${name}.txt"
    fi
done
