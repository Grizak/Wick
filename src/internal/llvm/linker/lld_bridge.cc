#include "lld/Common/Driver.h"
#include "llvm/ADT/ArrayRef.h"
#include "llvm/Support/raw_ostream.h"

LLD_HAS_DRIVER(coff)
LLD_HAS_DRIVER(elf)
LLD_HAS_DRIVER(macho)

namespace {
const lld::DriverDef drivers[] = {
    {lld::Gnu, &lld::elf::link},
    {lld::WinLink, &lld::coff::link},
    {lld::Darwin, &lld::macho::link},
};
}

extern "C" int Run_lld(int argc, const char *const *argv) {
    llvm::ArrayRef<const char *> args(argv, static_cast<size_t>(argc));
    lld::Result result =
        lld::lldMain(args, llvm::outs(), llvm::errs(), drivers);
    return result.retCode;
}
