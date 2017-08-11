#!/usr/bin/perl
use v5.14;
use warnings;

use Mojo::UserAgent;
use Mojo::DOM;

sub download_lyrics;
sub write_lyrics;
sub format_name;
sub list_files;

my $ua = Mojo::UserAgent->new(
	max_redirects => 5,
	connection_timeout => 10,
	inactivity_timeout => 10,
	request_timeout => 10,
	response_timeout => 10
);


# $ua->on(start => sub {
# 	my ($ua, $tx) = @_;
# 	$tx->req->headers->header('User-Agent' => 'Mozilla/5.0 (X11; Linux i686) ' .
# 		'AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.111 Safari/537.36');
# 	$tx->req->headers->header('Connection' => 'keep-alive');
# 	$tx->req->headers->header('Cache-Control' => 'max-age=0');
# 	$tx->req->headers->header('Accept' =>
# 	'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8');
# 	$tx->req->headers->header('Accept-Encoding' => 'gzip, deflate, sdch');
# 	$tx->req->headers->header('Accept-Language' => 'en-US,en;q=0.8,ru;q=0.6');
# });


my $artist;
my $song;

if (@ARGV == 0) {       # no arguments, read mocp info
	my $mocp_info = `mocp -i`;
	($artist) = $mocp_info =~ /artist: (.+?)\s*\n/i;
	($song)   = $mocp_info =~ /songtitle: (.+?)\s*\n/i;
	if (!defined $artist || !defined $song) {
		die "Couldn't read mocp info";
	}
} elsif (@ARGV == 1) {      # one argument passed, ex.: lyr "Led Zeppelin - Rock and Roll"
	($artist, $song) = $ARGV[0] =~ /(.+)-(.+)/;
} elsif (@ARGV == 2) {      #two args, from mocp or cli: lyr "pixies" "velouria"
	($artist, $song) = @ARGV;
} else {
	say "Usage: \nlyr\nlyr \"artist - song\"\nlyr artist song";
	exit;
}


$artist = lc $artist;
$song   = lc $song;

my $title = "$artist - $song";     #pretty title to store in text file
$title =~ s/([a-z]+)/\u$1/g;
($artist, $song) = format_name($artist, $song); #squashed name for filename and url 

my $fname = "$ENV{HOME}/.lyr/$artist-$song"; #name of file to store lyrics

if (! -e $fname) {
	my $text = download_lyrics($artist, $song);
	if ($text) {
		write_lyrics($fname, $title, $text);
	} else {
		die "$title lyrics not found";
	}
}
system "vim $fname";


sub write_lyrics {
	my ($fname, $title, $text) = @_;
	open my $file, '>', $fname;
	if (!$file) {
		warn "file $fname wasn't written";
		return;
	}
	print $file "$title\n\n";
	print $file $text;
}
	
sub format_name { #remove non-alpha-digit chars             
	my ($artist, $song) = @_;
	return undef if (!defined $artist || !defined $song);
	$artist =~ s/\Athe //;
	$artist =~ s/[^a-z0-9]//g;
	$song =~ s/[^a-z0-9]//g;
	return $artist, $song;
}

sub download_lyrics { # get lyrics from azlyrics.com
	my ($artist, $song) = @_;	
	return undef if (!defined $artist || !defined $song);

	my $url = "http://www.azlyrics.com/lyrics/$artist/$song.html";
	my $tx = $ua->get($url);
	die $tx->error->{message} if $tx->error;
	my $div = $tx->res->dom->at('div[class=lyricsh] ~ div ~ div');
	return undef unless $div;
	my $lyr =  $div->all_text(0);
	$lyr =~ s/^\s*//;
	$lyr =~ s/\s*$//;
	$lyr =~ tr/\r//d;
	return $lyr;
}

sub list_files {
	my $f = shift @_;
	my @files;
	state %seen;

	if (!-d $f) {
		return $f;
	}


	opendir(my $dir, $f) or return ();
	chdir $f;

	my $wd = `pwd`;
	chomp $wd;
	return () if exists $seen{$wd};
	$seen{$wd} = 1;

	for my $file (readdir $dir) {
		if ($file =~ m{(?:\A|/)\.{1,2}\z}) {
			next;
		}
		if (-d $file) {
			push @files, list_files($file);
		} else {
			push @files, $file;
		}
	}
	chdir ".." if $f !~ m{(?:\A|/)\.\z};
	return @files;
}
