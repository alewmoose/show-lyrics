#!/usr/bin/env perl
use v5.10;
use strict;
use warnings;
use utf8;

my $cmus_stat = `cmus-remote -Q`;
exit $? if $?;

my ($artist) = $cmus_stat =~ /^tag\s+artist\s+(.*)\s*$/mi;
my ($title)  = $cmus_stat =~ /^tag\s+title\s+(.*)\s*$/mi;
die "No artist/title\n" unless $artist && $title;
my $pretty_title = "$artist - $title";

say $pretty_title;

$artist =~ s/\Athe //i;
for ($artist, $title) {
    $_ = lc $_;
    s/[^a-z0-9]//g;
}
$title =~ s/[^a-z0-9]//g;

my $fname = "$ENV{HOME}/.lyr/$artist-$title";
exec 'less', '-c', $fname if -f $fname;

require Mojo::UserAgent;
my $ua = Mojo::UserAgent->new(
	max_redirects => 5,
);

my $url = "http://www.azlyrics.com/lyrics/$artist/$title.html";
my $tx = $ua->get($url);
die $tx->error->{message}, "\n" if $tx->error;

my $lyr = $tx->res->dom->at('div[class=lyricsh] ~ div ~ div')->all_text(0);
$lyr =~ s/^\s*//;
$lyr =~ s/\s*$//;

open my $fh, '>:encoding(UTF-8)', $fname or die "$fname: $!\n";
$fh->say($pretty_title, "\n");
$fh->say($lyr);
close $fh;

exec 'less', '-c', $fname;
