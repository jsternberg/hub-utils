variable "DESTDIR" {
  default = "/usr/local"
}

variable "HOME" {
  default = "$HOME"
}

target "install" {
  args = {
    DESTDIR = "/usr/local"
  }
  targets = ["binaries"]
  platforms = ["${BAKE_LOCAL_PLATFORM}"]
  output = ["type=local,dest=${DESTDIR}/bin"]
}

target "install-local" {
  inherits = ["install"]
  output = ["type=local,dest=${HOME}/bin"]
}
