 class Tz < Formula
   desc "Show time zones based on user defined groups"
   homepage "https://github.com/DavidGamba/dgtools/tree/master/tz"
   url "https://github.com/DavidGamba/dgtools/archive/refs/tags/tz/v0.1.0.tar.gz"
   sha256 "1ffaae8225ef3d7e3fdcf61348d2fb2b100bcd959cfaeed5ddeef7038c844786"

   depends_on "go" => :build

   def install
     cd "tz" do
       system "go", "get"
       system "go", "build"
       bin.install "tz"
     end
   end

   test do
     assert_match /Use 'tz help <command>' for extra details/, shell_output("#{bin}/tz")
   end
 end
