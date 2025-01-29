class InitiaVanity < Formula
    desc "Vanity address generator for Initia blockchain"
    homepage "https://github.com/degenhousedefi/initia-vanity"
    
    version "1.0.0"
  
    on_macos do
      if Hardware::CPU.arm?
        url "https://github.com/degenhousedefi/initia-vanity/releases/download/v#{version}/initia-vanity-darwin-arm64"
        sha256 "REPLACE_WITH_ACTUAL_SHA256"
      else
        url "https://github.com/degenhousedefi/initia-vanity/releases/download/v#{version}/initia-vanity-darwin-amd64"
        sha256 "REPLACE_WITH_ACTUAL_SHA256"
      end
    end
  
    on_linux do
      if Hardware::CPU.intel?
        url "https://github.com/degenhousedefi/initia-vanity/releases/download/v#{version}/initia-vanity-linux-amd64"
        sha256 "REPLACE_WITH_ACTUAL_SHA256"
      end
    end
  
    def install
      bin.install Dir["*"].first => "initia-vanity"
    end
  
    test do
      system "#{bin}/initia-vanity", "--version"
    end
  end