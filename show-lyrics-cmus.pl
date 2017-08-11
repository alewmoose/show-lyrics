#!/usr/bin/env perl
use v5.10;
use strict;
use warnings;
use utf8;
use subs qw(
    get_cmus_status
    parse_cmus_status
);

# TODO:
# remove tabs
# use HTTP::Tiny
# parse HTML with regexps :)
# use pretty artist & title directory and file namse

my ($exit_status, $cmus_status) = get_cmus_status;
exit $exit_status if $exit_status;

my $song_info = parse_cmus_status $cmus_status;
my $artist    = $song_info->{artist};
my $title     = $song_info->{title};

my $pretty_title = "$artist - $title";

$artist =~ s/\Athe //i;
for ($artist, $title) {
    $_ = lc $_;
    s/[^a-z0-9]//g;
}
$title =~ s/[^a-z0-9]//g;

my $cache_dir = "$ENV{HOME}/.lyrics";
unless (-d $cache_dir) {
    mkdir $cache_dir or die "Failed to create $cache_dir: $!\n";
}
my $artist_dir = "$cache_dir/$artist";
unless (-d $artist_dir) {
    mkdir $artist_dir or die "Failed to create $artist_dir: $!\n";
}

my $fname = "$artist_dir/$title";
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

sub get_cmus_status {
    my $cmus_status = `cmus-remote -Q`;
    return ($? >> 8, $cmus_status);
}

sub parse_cmus_status {
    my ($cmus_status) = @_;
    my ($artist) = $cmus_status =~ /^tag\s+artist\s+(.*)\s*$/mi;
    my ($title)  = $cmus_status =~ /^tag\s+title\s+(.*)\s*$/mi;
    die "Failed to parse cmus status" unless $artist && $title;
    return {
        artist => $artist,
        title  => $title,
    };
}
