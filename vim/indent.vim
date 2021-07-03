" Vim indent file
" Language:          Venice
" Maintainer:        Ian Fisher
" Latest Revision:   3 July 2021
"
" Based on http://psy.swansea.ac.uk/staff/carter/Vim/vim_indent.htm
"   and /usr/share/vim/vim80/indent/rust.vim

if exists("b:did_indent")
	finish
endif
let b:did_indent = 1

" Use C indentation rules for now.
setlocal cindent
