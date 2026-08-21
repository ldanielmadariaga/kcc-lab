#!/bin/bash
# Rates over a given set of kinds, from a pipe-delimited score file.
# A resource counts as clean when it has no missing and no mismatched field.
SP=/tmp/claude-1000/-home-luks-kcc-lab/b5a7e7f5-735e-462c-86f6-270be0273624/scratchpad
rates() { # $1=scorefile $2=kindlist $3=label
  awk -F'|' -v label="$3" 'NR==FNR{want[$1]=1; next}
    $1=="OK" && want[$2] {
      n++
      sm+=$4; smi+=$5; smm+=$7
      rm+=$8; rmi+=$9; rmm+=$11
      om+=$12; omi+=$13; omm+=$15
      if ($5==0 && $7==0)  scleanN++
      if ($9==0 && $11==0) rcleanN++
      if ($13==0 && $15==0) ocleanN++
    }
    END {
      printf "%s  (n=%d)\n", label, n
      printf "  spec                 %5.1f%%   clean %d/%d\n", 100*sm/(sm+smi+smm), scleanN, n
      printf "  required             %5.1f%%   clean %d/%d\n", 100*rm/(rm+rmi+rmm), rcleanN, n
      printf "  status.observedState %5.1f%%   clean %d/%d\n", 100*om/(om+omi+omm), ocleanN, n
    }' "$2" "$1"
}
rates "$1" "$2" "$3"
