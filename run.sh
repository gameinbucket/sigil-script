if [ -f compiler ]; then
  rm compiler
fi
cd go
go build -o compiler
cd ..
mv go/compiler .
if [ -f compiler ]; then
  ./compiler $@
fi