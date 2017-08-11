#!/usr/bin/perl
use v5.14;
use warnings;
use Mojo::UserAgent;
use Mojo::IOLoop;
use Mojo::DOM;

my $ua = Mojo::UserAgent->new(
	max_redirects => 5,
	connection_timeout => 10,
	inactivity_timeout => 10,
	request_timeout => 10,
	response_timeout => 10
);

# http://www.lyricsmode.com/lyrics/m/morphine/the_night.html
# http://lyrics.wikia.com/Pixies:The_Holiday_Song

my $artist;
my $song;

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

if (@ARGV == 0) { # read mocp info
	my $mocp_info = `mocp -i 2>&1`;
	die 'mocp ', $mocp_info =~ tr/\n//rd, "\n"
		if $?;
	#say $mocp_info;
	if ($mocp_info =~ /^Artist: \s+? (.+)$ \s*? SongTitle: \s+? (.+)$/xmi) {
		($artist, $song) = ($1, $2);
	} else {
		die "error: couldn't read mocp info\n";
	}

	$artist = lc $artist;
	$song = lc $song;
	$artist =~ s/\Athe //;
	$artist =~ s/[^a-z0-9]//g;
	$song =~ s/[^a-z0-9]//g;

	say "$artist - $song";



	my $url = "http://www.azlyrics.com/lyrics/$artist/$song.html";
	say $url;

	my $tx = $ua->get($url);
	die $tx->error->{message} if $tx->error;

	my $lyr = $tx->res->dom->at('div[class=lyricsh] ~ div ~ div')->all_text(0);
	$lyr =~ s/^\s*//;
	$lyr =~ s/\s*$//;
	say '-' x 40;
	say $lyr;
	print '-' x 40;

	
}




