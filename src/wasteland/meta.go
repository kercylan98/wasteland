package wasteland

type Meta interface {
	Address() Address

	Zone() Zone

	Version() Version

	LaunchAt() Timestamp

	Gen() Gen
}

func newMeta(address Address, zone Zone, version Version, launchAt Timestamp, gen Gen) Meta {
	return &metaImpl{
		address:  address,
		zone:     zone,
		version:  version,
		launchAt: launchAt,
		gen:      gen,
	}
}

type metaImpl struct {
	address  Address
	zone     Zone
	version  Version
	launchAt Timestamp
	gen      Gen
}

func (p *metaImpl) Address() Address {
	return p.address
}

func (p *metaImpl) Zone() Zone {
	return p.zone
}

func (p *metaImpl) Version() Version {
	return p.version
}

func (p *metaImpl) LaunchAt() Timestamp {
	return p.launchAt
}

func (p *metaImpl) Gen() Gen {
	return p.gen
}
